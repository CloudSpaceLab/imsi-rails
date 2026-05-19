const canvas = document.getElementById("route-field");
const ctx = canvas.getContext("2d", { alpha: false });

let width = 0;
let height = 0;
let nodes = [];
let raf = 0;

const colors = ["#3be4ac", "#f1b84b", "#53c7e8", "#a98bff", "#ff6b5f"];

function resize() {
  const ratio = Math.min(window.devicePixelRatio || 1, 2);
  width = window.innerWidth;
  height = window.innerHeight;
  canvas.width = Math.floor(width * ratio);
  canvas.height = Math.floor(height * ratio);
  canvas.style.width = `${width}px`;
  canvas.style.height = `${height}px`;
  ctx.setTransform(ratio, 0, 0, ratio, 0, 0);

  const count = width < 720 ? 36 : 58;
  nodes = Array.from({ length: count }, (_, index) => ({
    x: Math.random() * width,
    y: Math.random() * height,
    vx: (Math.random() - 0.5) * 0.22,
    vy: (Math.random() - 0.5) * 0.22,
    r: index % 7 === 0 ? 2.4 : 1.5,
    color: colors[index % colors.length],
  }));
}

function draw() {
  ctx.fillStyle = "#08090c";
  ctx.fillRect(0, 0, width, height);

  ctx.globalAlpha = 0.16;
  ctx.strokeStyle = "#fff8eb";
  ctx.lineWidth = 1;

  for (let i = 0; i < nodes.length; i += 1) {
    const a = nodes[i];
    for (let j = i + 1; j < nodes.length; j += 1) {
      const b = nodes[j];
      const dx = a.x - b.x;
      const dy = a.y - b.y;
      const dist = Math.sqrt(dx * dx + dy * dy);
      if (dist < 150) {
        ctx.globalAlpha = (150 - dist) / 800;
        ctx.beginPath();
        ctx.moveTo(a.x, a.y);
        ctx.lineTo(b.x, b.y);
        ctx.stroke();
      }
    }
  }

  for (const node of nodes) {
    node.x += node.vx;
    node.y += node.vy;

    if (node.x < -20) node.x = width + 20;
    if (node.x > width + 20) node.x = -20;
    if (node.y < -20) node.y = height + 20;
    if (node.y > height + 20) node.y = -20;

    ctx.globalAlpha = 0.86;
    ctx.fillStyle = node.color;
    ctx.beginPath();
    ctx.arc(node.x, node.y, node.r, 0, Math.PI * 2);
    ctx.fill();
  }

  raf = requestAnimationFrame(draw);
}

function rotateMetrics() {
  const health = document.getElementById("health-score");
  const credit = document.getElementById("credit-time");
  const degraded = document.getElementById("degraded-routes");
  if (!health || !credit || !degraded) return;

  const healthValue = 96.8 + Math.random() * 1.1;
  const creditValue = 38 + Math.floor(Math.random() * 9);
  const degradedValue = Math.random() > 0.62 ? 1 : 2;

  health.textContent = `${healthValue.toFixed(1)}%`;
  credit.textContent = `${creditValue}s`;
  degraded.textContent = String(degradedValue);
}

window.addEventListener("resize", () => {
  cancelAnimationFrame(raf);
  resize();
  draw();
});

resize();
draw();
setInterval(rotateMetrics, 2600);

