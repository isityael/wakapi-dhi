const chokidar = require("chokidar");
const { spawn } = require("node:child_process");

const buildCommand = process.argv.slice(2).join(" ").trim();

if (!buildCommand) {
  console.error("usage: node scripts/watch.js \"<build command>\"");
  process.exit(1);
}

const watchPaths = [
  "./views/**/*.html",
  "./static/assets/js/**/*.js",
  "./static/assets/css/**/*.css",
];

const ignored = ["**/vendor/**", "**/*.dist.*"];

let running = false;
let queued = false;

function runBuild() {
  if (running) {
    queued = true;
    return;
  }

  running = true;
  const child = spawn(buildCommand, { shell: true, stdio: "inherit" });

  child.on("exit", (code) => {
    running = false;
    if (queued) {
      queued = false;
      runBuild();
    } else if (code !== 0) {
      console.error(`[watch] build exited with code ${code}`);
    }
  });
}

chokidar
  .watch(watchPaths, {
    ignoreInitial: true,
    ignored,
  })
  .on("all", (event, path) => {
    console.log(`[watch] ${event}: ${path}`);
    runBuild();
  });

console.log("[watch] waiting for changes");
