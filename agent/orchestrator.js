#!/usr/bin/env node
/**
 * M2Sim Multi-Agent Orchestrator (Standalone)
 * Runs all agents sequentially in one cycle, then waits before next cycle.
 * No OpenClaw dependency.
 */

import { spawn, execSync } from 'child_process';
import { existsSync, mkdirSync, appendFileSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const REPO_DIR = join(__dirname, '..');
const SKILL_PATH = join(__dirname, 'skills');
const LOGS_DIR = join(__dirname, 'logs');
const TRACKER_ISSUE = 45;
const CYCLE_INTERVAL_MS = 600_000; // 10 minutes between cycles
const AGENT_TIMEOUT_MS = 900_000; // 15 minutes per agent
const MODEL = 'claude-opus-4-5';

// Agent sequence for each cycle
const AGENTS = ['alice', 'eric', 'bob', 'cathy', 'dana'];

// Track currently running agent
let currentAgentProcess = null;
let currentAgentName = null;
let cycleCount = 0;

function log(message) {
  const timestamp = new Date().toISOString().replace('T', ' ').slice(0, 19);
  console.log(`[${timestamp}] ${message}`);
}

function exec(cmd, options = {}) {
  try {
    return execSync(cmd, { 
      cwd: REPO_DIR, 
      encoding: 'utf-8',
      stdio: ['pipe', 'pipe', 'pipe'],
      ...options 
    }).trim();
  } catch (e) {
    return e.stdout?.trim() || '';
  }
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function runAgent(agent) {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
  const logFile = join(LOGS_DIR, `${agent}-${timestamp}.log`);
  
  log(`Running agent: ${agent}`);
  
  // Pull latest before each agent
  exec('git pull --rebase --quiet');
  
  const prompt = `You are [${agent}] working on the M2Sim project.

**Config:**
- GitHub Repo: sarchlab/m2sim  
- Local Path: ${REPO_DIR}
- Tracker Issue: #${TRACKER_ISSUE}

**Instructions:**
1. First, read the shared rules from: ${join(SKILL_PATH, 'everyone.md')}
2. Then read your specific role from: ${join(SKILL_PATH, `${agent}.md`)}
3. Execute your full cycle as described in your role file.
4. All GitHub activity (commits, PRs, comments) must start with [${agent}]

Work autonomously. Complete your tasks, then exit.`;

  return new Promise((resolve) => {
    const proc = spawn('claude', [
      '--model', MODEL,
      '--dangerously-skip-permissions',
      prompt
    ], {
      cwd: REPO_DIR,
      stdio: ['ignore', 'pipe', 'pipe']
    });

    currentAgentProcess = proc;
    currentAgentName = agent;

    const timeout = setTimeout(() => {
      log(`Agent ${agent} timed out, killing...`);
      proc.kill('SIGTERM');
    }, AGENT_TIMEOUT_MS);

    proc.stdout.on('data', (data) => {
      const text = data.toString();
      process.stdout.write(text);
      appendFileSync(logFile, text);
    });

    proc.stderr.on('data', (data) => {
      const text = data.toString();
      process.stderr.write(text);
      appendFileSync(logFile, text);
    });

    proc.on('close', (code) => {
      clearTimeout(timeout);
      currentAgentProcess = null;
      currentAgentName = null;
      log(`Agent ${agent} finished with code ${code}`);
      resolve(code);
    });

    proc.on('error', (err) => {
      clearTimeout(timeout);
      currentAgentProcess = null;
      currentAgentName = null;
      log(`Agent ${agent} error: ${err.message}`);
      resolve(1);
    });
  });
}

async function runGrace() {
  log('Running Grace (advisor)');
  await runAgent('grace');
}

async function runCycle() {
  cycleCount++;
  log(`========== CYCLE ${cycleCount} START ==========`);
  
  // Run Grace every 10th cycle
  if (cycleCount % 10 === 0) {
    await runGrace();
  }
  
  // Run all agents sequentially
  for (const agent of AGENTS) {
    await runAgent(agent);
  }
  
  log(`========== CYCLE ${cycleCount} END ==========`);
}

async function main() {
  log('M2Sim Orchestrator started (Node.js, standalone)');
  log(`Cycle interval: ${CYCLE_INTERVAL_MS / 1000}s, Agents: ${AGENTS.join(' → ')}`);
  log(`Repo: ${REPO_DIR}, Model: ${MODEL}`);
  
  // Create logs directory
  if (!existsSync(LOGS_DIR)) {
    mkdirSync(LOGS_DIR, { recursive: true });
  }
  
  // Main loop
  while (true) {
    await runCycle();
    
    log(`Sleeping ${CYCLE_INTERVAL_MS / 1000}s until next cycle...`);
    await sleep(CYCLE_INTERVAL_MS);
  }
}

// Handle graceful shutdown
process.on('SIGINT', () => {
  log('Shutting down...');
  if (currentAgentProcess) {
    log(`Killing agent ${currentAgentName}...`);
    currentAgentProcess.kill('SIGTERM');
  }
  process.exit(0);
});

process.on('SIGTERM', () => {
  log('Shutting down...');
  if (currentAgentProcess) {
    log(`Killing agent ${currentAgentName}...`);
    currentAgentProcess.kill('SIGTERM');
  }
  process.exit(0);
});

main().catch(console.error);
