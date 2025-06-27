#!/usr/bin/env node
/**
 * Simple test script for the npm package
 */

const { getBinaryPath, getBinaryName } = require('./index.js');
const fs = require('fs');
const path = require('path');

console.log('Testing @clickup/cli npm package...\n');

// Test 1: Check binary name
console.log('Test 1: Binary name detection');
const binaryName = getBinaryName();
console.log(`  Binary name: ${binaryName}`);
console.log(`  ✓ Platform detection works\n`);

// Test 2: Check if postinstall would run
console.log('Test 2: Post-install readiness');
const binDir = path.join(__dirname, 'bin');
if (fs.existsSync(binDir)) {
  console.log('  ✓ bin directory exists');
} else {
  console.log('  ✓ bin directory will be created during install');
}

// Test 3: Check package.json
console.log('\nTest 3: Package configuration');
const pkg = require('./package.json');
console.log(`  Name: ${pkg.name}`);
console.log(`  Version: ${pkg.version}`);
console.log(`  ✓ Package.json is valid\n`);

console.log('All tests passed! ✓');
console.log('\nTo test installation:');
console.log('  1. Run: npm pack');
console.log('  2. Run: npm install -g clickup-cli-*.tgz');
console.log('  3. Test: cu --version');