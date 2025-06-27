#!/usr/bin/env node
/**
 * ClickUp CLI npm package entry point
 * This file handles the execution of the cu binary
 */

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

// Determine the binary name based on platform
const getBinaryName = () => {
  const platform = process.platform;
  return platform === 'win32' ? 'cu.exe' : 'cu';
};

// Get the path to the binary
const getBinaryPath = () => {
  const binaryName = getBinaryName();
  const binPath = path.join(__dirname, 'bin', binaryName);
  
  if (!fs.existsSync(binPath)) {
    console.error('ClickUp CLI binary not found!');
    console.error('Please run: npm install -g @clickup/cli');
    console.error('');
    console.error('If the problem persists, please report it at:');
    console.error('https://github.com/timimsms/cu/issues');
    process.exit(1);
  }
  
  return binPath;
};

// Main execution
const main = () => {
  const binaryPath = getBinaryPath();
  const args = process.argv.slice(2);
  
  // Spawn the binary with inherited stdio
  const child = spawn(binaryPath, args, {
    stdio: 'inherit',
    shell: false
  });
  
  // Handle exit
  child.on('exit', (code) => {
    process.exit(code);
  });
  
  // Handle errors
  child.on('error', (err) => {
    if (err.code === 'ENOENT') {
      console.error('ClickUp CLI binary not found!');
      console.error('Path:', binaryPath);
    } else if (err.code === 'EACCES') {
      console.error('Permission denied when trying to execute ClickUp CLI');
      console.error('Try running: chmod +x', binaryPath);
    } else {
      console.error('Failed to start ClickUp CLI:', err.message);
    }
    process.exit(1);
  });
};

// Run if called directly
if (require.main === module) {
  main();
}

module.exports = { getBinaryPath, getBinaryName };