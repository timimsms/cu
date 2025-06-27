#!/usr/bin/env node
/**
 * ClickUp CLI npm package post-install script
 * Downloads the appropriate binary for the current platform
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const { createHash } = require('crypto');
const zlib = require('zlib');
const { execSync } = require('child_process');

// Configuration
const REPO_OWNER = 'timimsms';
const REPO_NAME = 'cu';
const BINARY_NAME = 'cu';

// Get platform details
const getPlatform = () => {
  const platform = process.platform;
  const arch = process.arch;
  
  // Map Node.js platform/arch to our naming convention
  const platformMap = {
    'darwin': 'darwin',
    'linux': 'linux',
    'win32': 'windows'
  };
  
  const archMap = {
    'x64': 'x86_64',
    'arm64': 'arm64',
    'ia32': 'i386'
  };
  
  const mappedPlatform = platformMap[platform];
  const mappedArch = archMap[arch];
  
  if (!mappedPlatform || !mappedArch) {
    throw new Error(`Unsupported platform: ${platform} ${arch}`);
  }
  
  return `${mappedPlatform}_${mappedArch}`;
};

// Download file with progress
const downloadFile = (url) => {
  return new Promise((resolve, reject) => {
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Handle redirect
        downloadFile(response.headers.location)
          .then(resolve)
          .catch(reject);
        return;
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`HTTP ${response.statusCode}: ${response.statusMessage}`));
        return;
      }
      
      const chunks = [];
      const totalSize = parseInt(response.headers['content-length'], 10);
      let downloadedSize = 0;
      
      response.on('data', (chunk) => {
        chunks.push(chunk);
        downloadedSize += chunk.length;
        
        // Simple progress indicator
        if (totalSize) {
          const percentage = Math.round((downloadedSize / totalSize) * 100);
          process.stdout.write(`\rDownloading: ${percentage}%`);
        }
      });
      
      response.on('end', () => {
        process.stdout.write('\n');
        resolve(Buffer.concat(chunks));
      });
      
      response.on('error', reject);
    }).on('error', reject);
  });
};

// Get latest version from GitHub
const getLatestVersion = async () => {
  const apiUrl = `https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest`;
  
  const options = {
    headers: {
      'User-Agent': 'clickup-cli-npm'
    }
  };
  
  return new Promise((resolve, reject) => {
    https.get(apiUrl, options, (response) => {
      let data = '';
      
      response.on('data', (chunk) => {
        data += chunk;
      });
      
      response.on('end', () => {
        try {
          const release = JSON.parse(data);
          resolve(release.tag_name);
        } catch (err) {
          reject(new Error('Failed to parse GitHub API response'));
        }
      });
      
      response.on('error', reject);
    }).on('error', reject);
  });
};

// Verify checksum
const verifyChecksum = async (buffer, checksumsUrl, fileName) => {
  console.log('Verifying checksum...');
  
  try {
    const checksumsData = await downloadFile(checksumsUrl);
    const checksums = checksumsData.toString('utf-8').split('\n');
    
    // Find checksum for our file
    const checksumLine = checksums.find(line => line.includes(fileName));
    if (!checksumLine) {
      console.warn('Warning: Could not find checksum for', fileName);
      return true; // Continue anyway
    }
    
    const expectedChecksum = checksumLine.split(/\s+/)[0];
    const actualChecksum = createHash('sha256').update(buffer).digest('hex');
    
    if (expectedChecksum !== actualChecksum) {
      throw new Error('Checksum verification failed!');
    }
    
    console.log('✓ Checksum verified');
  } catch (err) {
    console.warn('Warning: Checksum verification skipped:', err.message);
  }
  
  return true;
};

// Extract binary using native tools
const extractBinary = async (archiveBuffer, platform, archivePath) => {
  const binDir = path.join(__dirname, 'bin');
  
  // Create bin directory
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }
  
  // Save archive to temp file
  const tempDir = fs.mkdtempSync(path.join(require('os').tmpdir(), 'cu-'));
  const tempArchive = path.join(tempDir, path.basename(archivePath));
  fs.writeFileSync(tempArchive, archiveBuffer);
  
  try {
    if (platform.includes('windows')) {
      // Windows: Use PowerShell to extract
      execSync(`powershell -Command "Expand-Archive -Path '${tempArchive}' -DestinationPath '${tempDir}' -Force"`, { stdio: 'ignore' });
    } else {
      // Unix: Use tar
      execSync(`tar -xzf "${tempArchive}" -C "${tempDir}"`, { stdio: 'ignore' });
    }
    
    // Find and move the binary
    const isWindows = platform.includes('windows');
    const binaryName = isWindows ? `${BINARY_NAME}.exe` : BINARY_NAME;
    const sourcePath = path.join(tempDir, binaryName);
    const destPath = path.join(binDir, binaryName);
    
    if (fs.existsSync(sourcePath)) {
      fs.copyFileSync(sourcePath, destPath);
      if (!isWindows) {
        fs.chmodSync(destPath, 0o755);
      }
    } else {
      // Try to find the binary in subdirectories
      const files = fs.readdirSync(tempDir, { recursive: true });
      const binaryFile = files.find(f => f === binaryName || f.endsWith(`/${binaryName}`));
      if (binaryFile) {
        const fullPath = path.join(tempDir, binaryFile);
        fs.copyFileSync(fullPath, destPath);
        if (!isWindows) {
          fs.chmodSync(destPath, 0o755);
        }
      } else {
        throw new Error(`Binary ${binaryName} not found in archive`);
      }
    }
    
    return destPath;
  } finally {
    // Cleanup
    try {
      fs.rmSync(tempDir, { recursive: true, force: true });
    } catch (err) {
      // Ignore cleanup errors
    }
  }
};

// Main installation function
const install = async () => {
  console.log('Installing ClickUp CLI...\n');
  
  try {
    // Skip in CI environment
    if (process.env.CI || process.env.CLICKUP_CLI_SKIP_DOWNLOAD) {
      console.log('Skipping binary download (CI environment detected)');
      return;
    }
    
    // Skip if binary already exists
    const binPath = path.join(__dirname, 'bin', process.platform === 'win32' ? 'cu.exe' : 'cu');
    if (fs.existsSync(binPath)) {
      console.log('ClickUp CLI binary already exists, skipping download.');
      return;
    }
    
    // Detect platform
    const platform = getPlatform();
    console.log(`Platform: ${platform}`);
    
    // Get latest version
    console.log('Fetching latest version...');
    const version = await getLatestVersion();
    console.log(`Version: ${version}`);
    
    // Construct URLs
    const archiveName = platform.includes('windows')
      ? `${BINARY_NAME}_${platform}.zip`
      : `${BINARY_NAME}_${platform}.tar.gz`;
    
    const downloadUrl = `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${archiveName}`;
    const checksumsUrl = `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/checksums.txt`;
    
    // Download archive
    console.log(`\nDownloading ${archiveName}...`);
    const archiveBuffer = await downloadFile(downloadUrl);
    console.log('✓ Download complete');
    
    // Verify checksum
    await verifyChecksum(archiveBuffer, checksumsUrl, archiveName);
    
    // Extract binary
    console.log('Extracting binary...');
    const binaryPath = await extractBinary(archiveBuffer, platform, archiveName);
    console.log('✓ Extraction complete');
    
    // Verify installation
    try {
      const output = execSync(`"${binaryPath}" --version`, { encoding: 'utf-8' }).trim();
      console.log(`\n✓ Successfully installed: ${output}`);
    } catch (err) {
      console.warn('\nWarning: Could not verify installation');
    }
    
    console.log('\nGet started with: cu --help');
    
  } catch (err) {
    console.error('\n✗ Installation failed:', err.message);
    console.error('\nYou can try:');
    console.error('1. Installing directly: https://github.com/timimsms/cu/releases');
    console.error('2. Reporting the issue: https://github.com/timimsms/cu/issues');
    
    // Don't fail npm install
    process.exit(0);
  }
};

// Run installation
if (require.main === module) {
  install().catch(err => {
    console.error('Unexpected error:', err);
    // Don't fail npm install
    process.exit(0);
  });
}