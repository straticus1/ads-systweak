document.addEventListener('DOMContentLoaded', () => {
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
});
