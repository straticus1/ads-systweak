document.addEventListener('DOMContentLoaded', () => {
    // Shared state for new tabs (declared early so tab-switch handler can reference them)
    let autorunsData = [];
    let tcpLiveInterval = null;
    let allProcesses = [];
    const tcpLiveToggle = document.getElementById('tcpview-live');

    // Tab switching
    const navItems = document.querySelectorAll('.nav-item');
    const tabPanes = document.querySelectorAll('.tab-pane');

    navItems.forEach(item => {
        item.addEventListener('click', () => {
            const target = item.dataset.tab;

            navItems.forEach(n => n.classList.remove('active'));
            tabPanes.forEach(t => t.classList.remove('active'));

            item.classList.add('active');
            document.getElementById(`${target}-tab`).classList.add('active');

            if (target === 'defaults' && domainsList.length === 0) {
                loadDomains();
            } else if (target === 'autoruns' && autorunsData.length === 0) {
                loadAutoruns();
            } else if (target === 'tcpview') {
                loadTCPView();
                if (tcpLiveToggle.checked && !tcpLiveInterval) {
                    tcpLiveInterval = setInterval(loadTCPView, 3000);
                }
            } else if (target === 'procexp' && allProcesses.length === 0) {
                loadProcesses();
            } else if (target === 'reliability') {
                loadReliability();
            }

            // Stop tcpview auto-refresh when leaving that tab
            if (target !== 'tcpview') {
                clearInterval(tcpLiveInterval);
                tcpLiveInterval = null;
            }
        });
    });

    // Toast
    function showToast(msg) {
        const toast = document.getElementById('toast');
        toast.textContent = msg;
        toast.classList.add('show');
        setTimeout(() => toast.classList.remove('show'), 3000);
    }

    // --- TWEAKS LOGIC ---
    let allTweaks = [];
    const tweaksList = document.getElementById('tweaks-list');
    const tweakSearch = document.getElementById('tweak-search');

    async function loadTweaks() {
        try {
            const res = await fetch('/api/tweaks');
            allTweaks = await res.json();
            renderTweaks(allTweaks);
        } catch (e) {
            console.error(e);
            showToast('Failed to load tweaks');
        }
    }

    function renderTweaks(tweaks) {
        tweaksList.innerHTML = '';
        tweaks.forEach(t => {
            const card = document.createElement('div');
            card.className = 'tweak-card';
            card.innerHTML = `
                <div class="tweak-header">
                    <div class="tweak-title">${t.name}</div>
                    <div class="tweak-tags">
                        <span class="tweak-category">${t.category}</span>
                        <span class="risk-badge risk-${t.riskLevel}">${t.riskLevel} Risk</span>
                    </div>
                </div>
                <div class="tweak-desc">${t.description}</div>
                <label class="switch">
                    <input type="checkbox" id="chk-${t.id}" ${t.applied ? 'checked' : ''}>
                    <span class="slider"></span>
                    <span class="status-lbl">${t.applied ? 'Enabled' : 'Disabled'}</span>
                </label>
            `;
            
            const checkbox = card.querySelector(`#chk-${t.id}`);
            const lbl = card.querySelector('.status-lbl');
            
            checkbox.addEventListener('change', async (e) => {
                const action = e.target.checked ? 'apply' : 'revert';
                
                if (e.target.checked && (t.riskLevel === 'Medium' || t.riskLevel === 'High')) {
                    const promptMsg = t.riskLevel === 'High' 
                        ? 'WARNING: This is a HIGH RISK modification that could affect system stability or security. Are you sure you want to proceed?'
                        : 'This is a MEDIUM RISK modification. Are you sure you want to proceed?';
                    
                    if (!confirm(promptMsg)) {
                        e.target.checked = false;
                        return;
                    }
                }
                
                lbl.textContent = 'Applying...';
                checkbox.disabled = true;
                
                try {
                    const res = await fetch(`/api/tweaks/${t.id}/${action}`, { method: 'POST' });
                    if (!res.ok) throw new Error(await res.text());
                    lbl.textContent = e.target.checked ? 'Enabled' : 'Disabled';
                    showToast(`Tweak ${e.target.checked ? 'Applied' : 'Reverted'} successfully`);
                } catch (err) {
                    showToast(`Failed: ${err.message}`);
                    e.target.checked = !e.target.checked; // undo UI
                    lbl.textContent = e.target.checked ? 'Enabled' : 'Disabled';
                } finally {
                    checkbox.disabled = false;
                }
            });

            tweaksList.appendChild(card);
        });
    }

    tweakSearch.addEventListener('input', (e) => {
        const q = e.target.value.toLowerCase();
        const filtered = allTweaks.filter(t => 
            t.name.toLowerCase().includes(q) || 
            t.description.toLowerCase().includes(q)
        );
        renderTweaks(filtered);
    });

    // --- DEFAULTS LOGIC ---
    let domainsList = [];
    let currentDomain = '';
    let currentKeys = {};

    const domainListEl = document.getElementById('domain-list');
    const domainSearch = document.getElementById('domain-search');
    const domainHeader = document.querySelector('#domain-header h3');
    const keysTableBody = document.querySelector('#keys-table tbody');
    const keySearch = document.getElementById('key-search');
    const btnAddKey = document.getElementById('btn-add-key');
    const btnRestoreAll = document.getElementById('btn-restore-all');

    if (btnRestoreAll) {
        btnRestoreAll.addEventListener('click', async () => {
            if (confirm("WARNING: This will restore all macOS defaults that have been modified via this tool from their backups. Are you absolutely sure?")) {
                try {
                    btnRestoreAll.textContent = "Restoring...";
                    btnRestoreAll.disabled = true;
                    const res = await fetch('/api/defaults/restore', { method: 'POST' });
                    if (!res.ok) throw new Error(await res.text());
                    showToast("Restored successfully. Please restart your Mac.");
                    if (currentDomain) loadDomainKeys(currentDomain);
                } catch (e) {
                    showToast("Failed to restore: " + e.message);
                } finally {
                    btnRestoreAll.textContent = "Restore All Backups";
                    btnRestoreAll.disabled = false;
                }
            }
        });
    }

    // Modal
    const modal = document.getElementById('add-key-modal');
    const modalDomain = document.getElementById('modal-domain');
    const modalKey = document.getElementById('modal-key');
    const modalType = document.getElementById('modal-type');
    const modalValue = document.getElementById('modal-value');
    const btnModalCancel = document.getElementById('btn-modal-cancel');
    const btnModalSave = document.getElementById('btn-modal-save');

    async function loadDomains() {
        try {
            const res = await fetch('/api/defaults/domains');
            domainsList = await res.json();
            renderDomains(domainsList);
        } catch (e) {
            domainListEl.innerHTML = '<li style="color:red">Failed to load</li>';
        }
    }

    function renderDomains(domains) {
        domainListEl.innerHTML = '';
        domains.forEach(d => {
            const li = document.createElement('li');
            li.textContent = d;
            if (d === currentDomain) li.classList.add('active');
            
            li.addEventListener('click', () => {
                document.querySelectorAll('#domain-list li').forEach(el => el.classList.remove('active'));
                li.classList.add('active');
                loadDomainKeys(d);
            });
            domainListEl.appendChild(li);
        });
    }

    domainSearch.addEventListener('input', (e) => {
        const q = e.target.value.toLowerCase();
        renderDomains(domainsList.filter(d => d.toLowerCase().includes(q)));
    });

    async function loadDomainKeys(domain) {
        currentDomain = domain;
        domainHeader.textContent = domain;
        btnAddKey.disabled = false;
        keySearch.disabled = false;
        keySearch.value = '';
        keysTableBody.innerHTML = '<tr><td colspan="4">Loading...</td></tr>';
        
        try {
            const res = await fetch(`/api/defaults/domain/${domain}`);
            if (!res.ok) throw new Error("Failed to load");
            currentKeys = await res.json();
            renderKeys(currentKeys);
        } catch (e) {
            keysTableBody.innerHTML = `<tr><td colspan="4" style="color:red">Failed to load keys or domain is empty</td></tr>`;
            currentKeys = {};
        }
    }

    function detectType(val) {
        if (typeof val === 'boolean') return 'bool';
        if (typeof val === 'number') return Number.isInteger(val) ? 'int' : 'float';
        if (typeof val === 'string') return 'string';
        if (Array.isArray(val)) return 'array';
        if (typeof val === 'object') return 'dict';
        return 'string';
    }

    function renderKeys(keysMap, filter = '') {
        keysTableBody.innerHTML = '';
        const entries = Object.entries(keysMap);
        
        if (entries.length === 0) {
            keysTableBody.innerHTML = '<tr class="empty-state"><td colspan="4">No keys found</td></tr>';
            return;
        }

        let count = 0;
        entries.forEach(([key, val]) => {
            if (filter && !key.toLowerCase().includes(filter.toLowerCase())) return;
            count++;
            
            const tr = document.createElement('tr');
            const type = detectType(val);
            let displayVal = val;
            
            if (type === 'dict' || type === 'array') {
                displayVal = JSON.stringify(val);
                if (displayVal.length > 50) displayVal = displayVal.substring(0, 50) + '...';
            }

            tr.innerHTML = `
                <td style="font-weight:500">${key}</td>
                <td><span class="code-val">${type}</span></td>
                <td style="max-width:300px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${displayVal}</td>
                <td>
                    <button class="btn edit-text" data-key="${key}" data-type="${type}" data-val='${JSON.stringify(val)}'>Edit</button>
                    <button class="btn danger-text" data-key="${key}">Delete</button>
                </td>
            `;
            keysTableBody.appendChild(tr);
        });

        if (count === 0) {
            keysTableBody.innerHTML = '<tr class="empty-state"><td colspan="4">No matches found</td></tr>';
        }

        // Attach listeners
        document.querySelectorAll('.edit-text').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const k = e.target.dataset.key;
                const t = e.target.dataset.type;
                const v = JSON.parse(e.target.dataset.val);
                openModal(k, t, v);
            });
        });

        document.querySelectorAll('.danger-text').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const k = e.target.dataset.key;
                if (confirm(`Delete key ${k} from ${currentDomain}?`)) {
                    try {
                        const res = await fetch(`/api/defaults/domain/${currentDomain}/key/${k}`, { method: 'DELETE' });
                        if (!res.ok) throw new Error();
                        showToast(`Deleted ${k}`);
                        loadDomainKeys(currentDomain);
                    } catch {
                        showToast('Failed to delete');
                    }
                }
            });
        });
    }

    keySearch.addEventListener('input', (e) => {
        renderKeys(currentKeys, e.target.value);
    });

    btnAddKey.addEventListener('click', () => {
        openModal('', 'string', '');
    });

    function openModal(key, type, val) {
        if (type === 'dict' || type === 'array') {
            showToast('Editing Dict/Array natively is not supported via this simple UI yet.');
            return;
        }
        
        modalDomain.value = currentDomain;
        modalKey.value = key;
        modalKey.disabled = key !== ''; // Disable key edit if modifying existing
        modalType.value = type;
        modalValue.value = val;
        
        modal.classList.add('active');
    }

    btnModalCancel.addEventListener('click', () => {
        modal.classList.remove('active');
    });

    btnModalSave.addEventListener('click', async () => {
        const key = modalKey.value.trim();
        const type = modalType.value;
        const rawVal = modalValue.value;

        if (!key) return showToast('Key is required');

        let val = rawVal;
        if (type === 'bool') val = (rawVal.toLowerCase() === 'true' || rawVal === '1');
        if (type === 'int') val = parseInt(rawVal, 10);
        if (type === 'float') val = parseFloat(rawVal);

        btnModalSave.disabled = true;
        btnModalSave.textContent = 'Saving...';

        try {
            const res = await fetch('/api/defaults/key', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    domain: currentDomain,
                    key: key,
                    type: type,
                    value: val
                })
            });

            if (!res.ok) throw new Error(await res.text());
            
            showToast('Key saved successfully');
            modal.classList.remove('active');
            loadDomainKeys(currentDomain);
        } catch (e) {
            showToast(`Error: ${e.message}`);
        } finally {
            btnModalSave.disabled = false;
            btnModalSave.textContent = 'Save';
        }
    });

    // Start
    loadTweaks();

    // --- AUTORUNS ---
    async function loadAutoruns() {
        const tbody = document.querySelector('#autoruns-table tbody');
        tbody.innerHTML = '<tr class="empty-state"><td colspan="6">Loading...</td></tr>';
        try {
            const res = await fetch('/api/autoruns');
            autorunsData = await res.json();
            renderAutoruns(autorunsData);
        } catch(e) { tbody.innerHTML = '<tr class="empty-state"><td colspan="6">Failed to load</td></tr>'; }
    }

    function renderAutoruns(items) {
        const tbody = document.querySelector('#autoruns-table tbody');
        if (!items || items.length === 0) {
            tbody.innerHTML = '<tr class="empty-state"><td colspan="6">No startup items found</td></tr>';
            return;
        }
        tbody.innerHTML = '';
        items.forEach(item => {
            const tr = document.createElement('tr');
            const typeColor = { LaunchAgent:'#007aff', LaunchDaemon:'#ff9f0a', LoginItem:'#34c759', CronJob:'#af52de', KernelExtension:'#ff3b30' };
            const color = typeColor[item.type] || '#8a8a93';
            tr.innerHTML = `
                <td><span style="color:${color};font-size:11px;font-weight:600;background:${color}1a;padding:2px 6px;border-radius:4px">${item.type}</span></td>
                <td style="font-family:monospace;font-size:12px;max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${item.label}">${item.label}</td>
                <td style="font-family:monospace;font-size:12px;color:var(--text-secondary);max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${item.program}">${item.program || '—'}</td>
                <td style="font-size:12px;color:var(--text-secondary)">${item.source || '—'}</td>
                <td style="font-size:12px">${item.runAtLoad ? '<span style="color:#34c759">Yes</span>' : '<span style="color:var(--text-secondary)">No</span>'}</td>
                <td>${item.readOnly ? '<span style="font-size:12px;color:var(--text-secondary)">System</span>' :
                    `<button class="btn ${item.enabled ? 'danger-text' : 'edit-text'} autorun-toggle" data-label="${encodeURIComponent(item.label)}" data-enabled="${item.enabled}" style="font-size:12px;padding:3px 10px">${item.enabled ? 'Disable' : 'Enable'}</button>`
                }</td>
            `;
            tbody.appendChild(tr);
        });
        document.querySelectorAll('.autorun-toggle').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const label = decodeURIComponent(btn.dataset.label);
                const enabled = btn.dataset.enabled === 'true';
                const action = enabled ? 'disable' : 'enable';
                btn.disabled = true; btn.textContent = '...';
                try {
                    const res = await fetch(`/api/autoruns/${encodeURIComponent(label)}/${action}`, { method: 'POST' });
                    if (!res.ok) throw new Error(await res.text());
                    showToast(`${action === 'disable' ? 'Disabled' : 'Enabled'}: ${label}`);
                    loadAutoruns();
                } catch(e) { showToast(`Failed: ${e.message}`); loadAutoruns(); }
            });
        });
    }
    document.getElementById('btn-autoruns-refresh').addEventListener('click', loadAutoruns);

    // --- TCPVIEW ---
    async function loadTCPView() {
        try {
            const res = await fetch('/api/tcpview');
            const sockets = await res.json();
            const tbody = document.querySelector('#tcpview-table tbody');
            if (!sockets || sockets.length === 0) {
                tbody.innerHTML = '<tr class="empty-state"><td colspan="7">No sockets found</td></tr>';
                return;
            }
            tbody.innerHTML = '';
            const stateColor = { ESTABLISHED:'#34c759', LISTEN:'#007aff', CLOSE_WAIT:'#ff9f0a', TIME_WAIT:'#ff9f0a', CLOSED:'#8a8a93' };
            sockets.forEach(s => {
                const tr = document.createElement('tr');
                const color = stateColor[s.state] || '#8a8a93';
                tr.innerHTML = `
                    <td style="font-family:monospace;font-size:12px">${s.pid}</td>
                    <td style="font-weight:500">${s.processName}</td>
                    <td><span style="font-size:11px;color:#007aff;background:rgba(0,122,255,0.1);padding:2px 6px;border-radius:4px;font-family:monospace">${s.protocol}</span></td>
                    <td style="font-family:monospace;font-size:12px">${s.localAddr}</td>
                    <td style="font-family:monospace;font-size:12px;color:var(--text-secondary)">${s.remoteAddr || '—'}</td>
                    <td><span style="font-size:11px;color:${color};background:${color}1a;padding:2px 6px;border-radius:4px;font-weight:600">${s.state || '—'}</span></td>
                    <td style="font-size:12px;color:var(--text-secondary)">${s.user}</td>
                `;
                tbody.appendChild(tr);
            });
            document.getElementById('tcpview-status').textContent = `${sockets.length} sockets · ${new Date().toLocaleTimeString()}`;
        } catch(e) { document.querySelector('#tcpview-table tbody').innerHTML = '<tr class="empty-state"><td colspan="7">Failed to load</td></tr>'; }
    }

    tcpLiveToggle.addEventListener('change', () => {
        if (tcpLiveToggle.checked) {
            tcpLiveInterval = setInterval(loadTCPView, 3000);
        } else {
            clearInterval(tcpLiveInterval);
        }
    });

    // --- PROCESS EXPLORER ---
    async function loadProcesses() {
        const tbody = document.querySelector('#proc-table tbody');
        tbody.innerHTML = '<tr class="empty-state"><td colspan="7">Loading...</td></tr>';
        try {
            const res = await fetch('/api/processes');
            allProcesses = await res.json();
            renderProcesses(allProcesses);
        } catch(e) { tbody.innerHTML = '<tr class="empty-state"><td colspan="7">Failed to load</td></tr>'; }
    }

    function renderProcesses(procs) {
        const tbody = document.querySelector('#proc-table tbody');
        tbody.innerHTML = '';
        if (!procs || procs.length === 0) {
            tbody.innerHTML = '<tr class="empty-state"><td colspan="7">No processes</td></tr>';
            return;
        }
        procs.forEach(p => {
            const tr = document.createElement('tr');
            tr.style.cursor = 'pointer';
            const cpuColor = p.cpu > 50 ? '#ff3b30' : p.cpu > 20 ? '#ff9f0a' : 'inherit';
            const cmd = p.command || p.name || '';
            tr.innerHTML = `
                <td style="font-family:monospace;font-size:12px">${p.pid}</td>
                <td style="font-weight:500">${p.name}</td>
                <td style="font-size:12px;color:var(--text-secondary)">${p.user}</td>
                <td style="font-family:monospace;font-size:12px;color:${cpuColor}">${(p.cpu||0).toFixed(1)}%</td>
                <td style="font-family:monospace;font-size:12px">${(p.mem||0).toFixed(1)}%</td>
                <td style="font-family:monospace;font-size:12px;color:var(--text-secondary)">${p.state||'?'}</td>
                <td style="font-size:12px;color:var(--text-secondary);max-width:300px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${cmd}">${cmd}</td>
            `;
            tr.addEventListener('click', () => showProcessDetail(p));
            tbody.appendChild(tr);
        });
    }

    document.getElementById('proc-search').addEventListener('input', (e) => {
        const q = e.target.value.toLowerCase();
        renderProcesses(allProcesses.filter(p => p.name.toLowerCase().includes(q) || String(p.pid).includes(q) || (p.user||'').toLowerCase().includes(q)));
    });

    async function showProcessDetail(p) {
        const panel = document.getElementById('proc-detail-panel');
        document.getElementById('proc-detail-title').textContent = `${p.name} (PID ${p.pid})`;
        document.getElementById('proc-detail-content').innerHTML = '<div style="color:var(--text-secondary)">Loading...</div>';
        panel.style.display = 'block';
        try {
            const res = await fetch(`/api/processes/${p.pid}`);
            const detail = await res.json();
            const section = (title, items) => items && items.length > 0
                ? `<div style="margin-bottom:16px"><div style="font-weight:600;margin-bottom:6px;color:var(--text-primary)">${title} <span style="font-size:11px;color:var(--text-secondary)">(${items.length})</span></div>${items.slice(0,50).map(f=>`<div style="font-family:monospace;font-size:11px;color:var(--text-secondary);word-break:break-all;margin-bottom:2px">${f}</div>`).join('')}${items.length>50?`<div style="font-size:11px;color:var(--text-secondary)">...+${items.length-50} more</div>`:''}</div>`
                : '';
            document.getElementById('proc-detail-content').innerHTML =
                section('Open Files', detail.openFiles) +
                section('Dylibs', detail.dylibs) +
                section('Sockets', detail.sockets) +
                section('Environment', detail.envVars) ||
                '<div style="color:var(--text-secondary)">No additional details available (may require elevated permissions)</div>';
        } catch(e) {
            document.getElementById('proc-detail-content').innerHTML = '<div style="color:var(--danger)">Failed to load details</div>';
        }
    }

    document.getElementById('btn-proc-detail-close').addEventListener('click', () => {
        document.getElementById('proc-detail-panel').style.display = 'none';
    });
    document.getElementById('btn-proc-refresh').addEventListener('click', loadProcesses);

    // --- RELIABILITY MONITOR ---
    async function loadReliability() {
        const tbody = document.querySelector('#reliability-table tbody');
        tbody.innerHTML = '<tr class="empty-state"><td colspan="6">Loading...</td></tr>';
        try {
            const res = await fetch('/api/reliability');
            const events = await res.json();
            if (!events || events.length === 0) {
                tbody.innerHTML = '<tr class="empty-state"><td colspan="6">No crash/hang reports found</td></tr>';
                return;
            }
            tbody.innerHTML = '';
            const typeColor = { Crash:'#ff3b30', Hang:'#ff9f0a', Panic:'#ff3b30', Spin:'#af52de', Wakeup:'#007aff' };
            events.forEach(ev => {
                const tr = document.createElement('tr');
                tr.style.cursor = 'pointer';
                const color = typeColor[ev.type] || '#8a8a93';
                const ts = ev.timestamp ? new Date(ev.timestamp).toLocaleString() : '—';
                tr.innerHTML = `
                    <td style="font-size:12px;color:var(--text-secondary);white-space:nowrap">${ts}</td>
                    <td><span style="font-size:11px;color:${color};background:${color}1a;padding:2px 6px;border-radius:4px;font-weight:600">${ev.type}</span></td>
                    <td style="font-weight:500">${ev.appName}</td>
                    <td style="font-size:12px;color:var(--text-secondary)">${ev.version||'—'}</td>
                    <td style="font-size:12px;color:var(--text-secondary)">${ev.os||'—'}</td>
                    <td style="font-size:12px;color:var(--text-secondary);max-width:300px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${ev.summary||''}">${ev.summary||'—'}</td>
                `;
                tr.addEventListener('click', () => {
                    document.getElementById('reliability-detail-title').textContent = `${ev.appName} — ${ev.type} at ${ts}`;
                    document.getElementById('reliability-detail-body').textContent = ev.summary || '(no summary)';
                    document.getElementById('reliability-detail').style.display = 'block';
                });
                tbody.appendChild(tr);
            });
        } catch(e) { tbody.innerHTML = '<tr class="empty-state"><td colspan="6">Failed to load</td></tr>'; }
    }

    document.getElementById('btn-reliability-refresh').addEventListener('click', loadReliability);
    document.getElementById('btn-reliability-close').addEventListener('click', () => {
        document.getElementById('reliability-detail').style.display = 'none';
    });
});
