const status = (msg) => { document.getElementById('status').textContent = msg; };

let currentYaml = '';

async function boot() {
  const go = new Go();
  const resp = await fetch('wg-manage.wasm');
  const bytes = await resp.arrayBuffer();
  const result = await WebAssembly.instantiate(bytes, go.importObject);
  go.run(result.instance);
  await new Promise(r => setTimeout(r, 300));
  if (typeof wgManage === 'undefined') {
    status('wasm API not found');
    return;
  }
  status('wasm ready ✔ (wg-manage ' + wgManage.version().result + ')');
  document.getElementById('app').hidden = false;

  const yamlEl = document.getElementById('yaml');
  const saved = localStorage.getItem('wg-manage-yaml');
  if (saved) {
    yamlEl.value = saved;
    currentYaml = saved;
    refreshPeers();
  }

  document.getElementById('btnBootstrap').onclick = () => {
    const endpoint = prompt('Server endpoint (host:port):', 'vpn.example.com:51820');
    if (!endpoint) return;
    const res = wgManage.bootstrap(endpoint);
    if (!res.ok) { status('error: ' + res.error); return; }
    setYaml(res.result);
  };

  document.getElementById('btnParse').onclick = () => {
    const res = wgManage.parseYaml(yamlEl.value);
    status(res.ok ? 'YAML valid ✔ (normalized below)' : 'invalid: ' + res.error);
    if (res.ok) setYaml(res.result);
  };

  document.getElementById('btnViz').onclick = () => {
    const res = wgManage.viz(yamlEl.value, 'ascii');
    const out = document.getElementById('vizOut');
    out.hidden = false;
    out.textContent = res.ok ? res.result : 'error: ' + res.error;
  };

  yamlEl.oninput = () => {
    currentYaml = yamlEl.value;
    localStorage.setItem('wg-manage-yaml', currentYaml);
    refreshPeers();
  };
}

function setYaml(y) {
  currentYaml = y;
  document.getElementById('yaml').value = y;
  localStorage.setItem('wg-manage-yaml', y);
  refreshPeers();
  status('updated');
}

function refreshPeers() {
  const res = wgManage.peerNames(currentYaml);
  const el = document.getElementById('peers');
  el.innerHTML = '';
  if (!res.ok) { status('invalid yaml: ' + res.error); return; }
  for (let i = 0; i < res.result.length; i++) {
    const p = res.result.get(i);
    const div = document.createElement('div');
    div.className = 'peer';
    const badge = p.server ? '<span class="badge server">server</span>' : '<span class="badge">client</span>';
    div.innerHTML = `<span><b>${escapeHtml(p.name)}</b> ${escapeHtml(p.address)} ${p.endpoint ? escapeHtml(p.endpoint) : ''}</span>${badge}`;
    div.onclick = () => {
      const c = wgManage.generateConf(currentYaml, p.name);
      const view = document.getElementById('confView');
      view.style.display = 'block';
      document.getElementById('conf').textContent = c.ok ? c.result : 'error: ' + c.error;
    };
    el.appendChild(div);
  }
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
}

if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('sw.js');
}

boot();
