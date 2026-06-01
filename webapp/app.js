const tg = window.Telegram && window.Telegram.WebApp;
        if (tg) {
            tg.ready();
            tg.expand();
            document.documentElement.style.setProperty('--tg-theme-bg-color', tg.backgroundColor || '#e5ddd5');
            document.documentElement.style.setProperty('--tg-theme-text-color', tg.textColor || '#111');
            document.documentElement.style.setProperty('--tg-theme-hint-color', tg.hintColor || '#706f6f');
            document.documentElement.style.setProperty('--tg-theme-button-color', tg.buttonColor || '#2aabee');
            document.documentElement.style.setProperty('--tg-theme-button-text-color', tg.buttonTextColor || '#fff');
            document.documentElement.style.setProperty('--tg-theme-secondary-bg-color', tg.secondaryBackgroundColor || '#fff');
            if (tg.themeParams && tg.themeParams.section_header_text_color) {
                document.documentElement.style.setProperty('--tg-theme-header-text-color', tg.themeParams.section_header_text_color);
            }
            if (tg.themeParams && tg.themeParams.section_bg_color) {
                document.documentElement.style.setProperty('--tg-theme-header-bg-color', tg.themeParams.section_bg_color);
            }
            if (tg.MainButton) tg.MainButton.hide();
        }

        const STORAGE_KEY = 'grounded_llm_session_id';
        const DOMAIN_STORAGE_KEY = 'grounded_llm_domain_id';
        const LOCALE_STORAGE_KEY = 'grounded_llm_locale';
        const API_BASE_STORAGE_KEY = 'grounded_llm_api_base';
        const API_BASE_SCHEMA_VERSION = '2';

        if (sessionStorage.getItem('grounded_llm_api_base_v') !== API_BASE_SCHEMA_VERSION) {
            sessionStorage.removeItem(API_BASE_STORAGE_KEY);
            sessionStorage.setItem('grounded_llm_api_base_v', API_BASE_SCHEMA_VERSION);
        }

        let apiBaseUrl = sessionStorage.getItem(API_BASE_STORAGE_KEY) || '/api/';

        function detectLocale() {
            var stored = sessionStorage.getItem(LOCALE_STORAGE_KEY);
            if (stored === 'ru' || stored === 'en') return stored;
            if (tg && tg.initDataUnsafe && tg.initDataUnsafe.user && tg.initDataUnsafe.user.language_code) {
                var lc = String(tg.initDataUnsafe.user.language_code).toLowerCase();
                if (lc.indexOf('ru') === 0) return 'ru';
                if (lc.indexOf('en') === 0) return 'en';
            }
            var nav = (navigator.language || '').toLowerCase();
            if (nav.indexOf('ru') === 0) return 'ru';
            if (nav.indexOf('en') === 0) return 'en';
            return 'ru';
        }

        let uiLocale = detectLocale();

        let sessionId = null;
        let domainId = sessionStorage.getItem(DOMAIN_STORAGE_KEY) || 'default';
        let sending = false;

        const el = {
            messagesRoot: document.getElementById('messagesRoot'),
            chatScroll: document.getElementById('chatScroll'),
            inputText: document.getElementById('inputText'),
            sendBtn: document.getElementById('sendBtn'),
            typingLine: document.getElementById('typingLine'),
            toast: document.getElementById('toast'),
            domainSelect: document.getElementById('domainSelect'),
            onboardingRoot: document.getElementById('onboardingRoot'),
            onboardingChips: document.getElementById('onboardingChips'),
            headerTitle: document.getElementById('headerTitle'),
            headerSubtitle: document.getElementById('headerSubtitle'),
            domainLabel: document.getElementById('domainLabel'),
            headerDisclaimer: document.getElementById('headerDisclaimer'),
            onboardingTitle: document.getElementById('onboardingTitle'),
            chatDivider: document.getElementById('chatDivider'),
        };

        async function loadBranding() {
            try {
                var res = await apiFetch('/branding' + localeQuery(), { method: 'GET' });
                var data = parseApiResponseJson(await res.text());
                if (!data.success || !data.branding) return;
                var b = data.branding;
                if (el.headerTitle && b.app_title) {
                    el.headerTitle.textContent = (b.header_emoji ? b.header_emoji + ' ' : '') + b.app_title;
                }
                if (el.headerSubtitle && b.header_subtitle) el.headerSubtitle.textContent = b.header_subtitle;
                if (el.domainLabel && b.domain_label) el.domainLabel.textContent = b.domain_label;
                if (el.headerDisclaimer && b.disclaimer) el.headerDisclaimer.textContent = b.disclaimer;
                if (el.onboardingTitle && b.onboarding_title) el.onboardingTitle.textContent = b.onboarding_title;
                if (el.chatDivider && b.chat_divider) el.chatDivider.textContent = b.chat_divider;
                if (b.app_title) document.title = b.app_title + ' — чат';
            } catch (e) {
                console.warn('loadBranding', e);
            }
        }

        function showToast(msg) {
            el.toast.textContent = msg;
            el.toast.classList.add('show');
            clearTimeout(showToast._t);
            showToast._t = setTimeout(function() { el.toast.classList.remove('show'); }, 4200);
        }

        /** initData от Telegram — криптографически подписанные данные пользователя (см. core.telegram.org/bots/webapps). */
        function getTelegramInitData() {
            if (tg && tg.initData) {
                return String(tg.initData);
            }
            return '';
        }

        /** Заголовки для API: initData уходит на Go для проверки подписи ботом. */
        function withAuthHeaders(extra) {
            var h = Object.assign({}, extra || {});
            var initData = getTelegramInitData();
            if (initData) {
                h['X-Telegram-Init-Data'] = initData;
            }
            h['X-Locale'] = uiLocale;
            h['Accept-Language'] = uiLocale;
            return h;
        }

        function localeQuery() {
            return '?locale=' + encodeURIComponent(uiLocale);
        }

        function dedupeApiBases(list) {
            var out = [];
            var seen = {};
            for (var i = 0; i < list.length; i++) {
                var b = list[i];
                if (!b || seen[b]) continue;
                seen[b] = true;
                out.push(b);
            }
            return out;
        }

        /** Прямой Go на 8080 (обход nginx без proxy). */
        function alternateApiBase8080() {
            try {
                var p = window.location.protocol;
                var h = window.location.hostname;
                if (!h) return null;
                if (String(window.location.port) === '8080') return null;
                var bases = [];
                bases.push('http://127.0.0.1:8080/api/');
                if (h !== '127.0.0.1') {
                    bases.push(p + '//' + h + ':8080/api/');
                }
                return bases;
            } catch (e) {
                return null;
            }
        }

        /** Только ответы нашего Go: поле success. Иначе чужой JSON ({"error":"Not Found"}) обрывал перебор URL. */
        function isOurAPIJsonBody(txt) {
            var t = String(txt).trim();
            if (t.charAt(0) !== '{') return false;
            try {
                var o = JSON.parse(t);
                if (!o || typeof o !== 'object') return false;
                return Object.prototype.hasOwnProperty.call(o, 'success');
            } catch (e) {
                return false;
            }
        }

        function buildApiCandidates() {
            var port = String(window.location.port || '');
            var list = [];
            list.push(apiBaseUrl);
            list.push('/api/');
            var alts = alternateApiBase8080();
            if (alts) {
                for (var a = 0; a < alts.length; a++) {
                    list.push(alts[a]);
                }
            }
            return dedupeApiBases(list);
        }

        /**
         * Запрос к API: сначала тот же origin (/api/), затем Go на :8080.
         * path — с ведущим слэшем, напр. "/session" (итого /api/session).
         */
        async function apiFetch(path, init) {
            var candidates = buildApiCandidates();
            var lastRes = null;
            for (var i = 0; i < candidates.length; i++) {
                var base = candidates[i];
                var baseNorm = base.endsWith('/') ? base : base + '/';
                var pathNorm = String(path).replace(/^\//, '');
                var url = baseNorm + pathNorm;
                var res;
                try {
                    var opts = init ? Object.assign({}, init) : {};
                    opts.headers = withAuthHeaders(opts.headers);
                    if (!opts.signal && url.indexOf(':8080') !== -1 &&
                        typeof AbortSignal !== 'undefined' && typeof AbortSignal.timeout === 'function') {
                        opts.signal = AbortSignal.timeout(5000);
                    }
                    res = await fetch(url, opts);
                } catch (e) {
                    continue;
                }
                lastRes = res;
                var peek = await res.clone().text();
                if (res.ok || isOurAPIJsonBody(peek)) {
                    if (i > 0) {
                        apiBaseUrl = baseNorm;
                        sessionStorage.setItem(API_BASE_STORAGE_KEY, apiBaseUrl);
                    }
                    return res;
                }
            }
            if (!lastRes) {
                throw new Error('No API connection. Start docker compose (webapp + server) or Go on port 8080.');
            }
            return lastRes;
        }

        /**
         * Парсит JSON-объект из тела ответа. Gin при 404 отдаёт текст "404 page not found" —
         * тогда JSON.parse читает число 404 и падает на «position 4» (буква «p» в «page»).
         */
        function parseApiResponseJson(raw) {
            var s = String(raw).replace(/^\uFEFF/, '').trim();
            if (!s) {
                throw new Error('Empty server response');
            }
            if (s.indexOf('404 page not found') === 0 || /^404\s/.test(s)) {
                throw new Error('API route not found (404). Restart containers: docker compose up --build');
            }
            if (s.charAt(0) === '<') {
                throw new Error('Expected JSON but received HTML — check proxy and API URL.');
            }
            var i = s.indexOf('{');
            if (i < 0) {
                throw new Error('Response is not JSON: ' + s.slice(0, 200));
            }
            var depth = 0;
            var inStr = false;
            var esc = false;
            for (var j = i; j < s.length; j++) {
                var c = s[j];
                if (inStr) {
                    if (esc) {
                        esc = false;
                        continue;
                    }
                    if (c === '\\') {
                        esc = true;
                        continue;
                    }
                    if (c === '"') {
                        inStr = false;
                    }
                    continue;
                }
                if (c === '"') {
                    inStr = true;
                    continue;
                }
                if (c === '{') {
                    depth++;
                } else if (c === '}') {
                    depth--;
                    if (depth === 0) {
                        return JSON.parse(s.slice(i, j + 1));
                    }
                }
            }
            throw new Error('Incomplete JSON in server response');
        }

        async function loadOnboarding(selectedDomain) {
            try {
                var res = await apiFetch('/onboarding?domain_id=' + encodeURIComponent(selectedDomain || domainId) + '&locale=' + encodeURIComponent(uiLocale), { method: 'GET' });
                var data = parseApiResponseJson(await res.text());
                var questions = (data.success && data.questions) ? data.questions : [];
                el.onboardingChips.innerHTML = '';
                if (!questions.length) {
                    el.onboardingRoot.hidden = true;
                    return;
                }
                questions.forEach(function(q) {
                    var btn = document.createElement('button');
                    btn.type = 'button';
                    btn.className = 'onboarding-chip';
                    btn.textContent = q;
                    btn.addEventListener('click', function() {
                        el.inputText.value = q;
                        autoResize();
                        sendMessage();
                    });
                    el.onboardingChips.appendChild(btn);
                });
                updateOnboardingVisibility();
            } catch (e) {
                console.error('loadOnboarding', e);
                el.onboardingRoot.hidden = true;
            }
        }

        function updateOnboardingVisibility() {
            var hasMessages = el.messagesRoot.querySelector('.row');
            el.onboardingRoot.hidden = !el.onboardingChips.children.length || !!hasMessages;
        }

        async function sendFeedback(messageId, rating) {
            if (!sessionId || !messageId) return;
            try {
                var res = await apiFetch('/feedback', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json; charset=utf-8' },
                    body: JSON.stringify({ session_id: sessionId, message_id: messageId, rating: rating })
                });
                var data = parseApiResponseJson(await res.text());
                if (!res.ok || !data.success) {
                    showToast(data.error || 'Failed to save rating');
                    return;
                }
                var btn = el.messagesRoot.querySelector('[data-feedback-for="' + messageId + '"][data-rating="' + rating + '"]');
                if (btn && btn.parentElement) {
                    btn.parentElement.querySelectorAll('.feedback-btn').forEach(function(b) {
                        b.classList.toggle('active', Number(b.getAttribute('data-rating')) === rating);
                        b.disabled = true;
                    });
                }
            } catch (e) {
                showToast(e.message || 'Rating error');
            }
        }

        async function loadDomainsCatalog() {
            try {
                var res = await apiFetch('/domains' + localeQuery(), { method: 'GET' });
                var data = parseApiResponseJson(await res.text());
                if (!data.success || !data.domains) return;
                el.domainSelect.innerHTML = '';
                data.domains.forEach(function(c) {
                    var opt = document.createElement('option');
                    opt.value = c.id;
                    var label = (c.emoji ? c.emoji + ' ' : '') + (c.name || c.name_ru || c.id);
                    if (!c.rag_enabled) label += ' (скоро)';
                    opt.textContent = label;
                    el.domainSelect.appendChild(opt);
                });
                domainId = sessionStorage.getItem(DOMAIN_STORAGE_KEY) || data.default_domain || 'default';
                el.domainSelect.value = domainId;
            } catch (e) {
                console.error('loadDomainsCatalog', e);
            }
        }

        async function createSessionWithDomain(selectedDomain) {
            domainId = selectedDomain;
            sessionStorage.setItem(DOMAIN_STORAGE_KEY, domainId);
            sessionStorage.removeItem(STORAGE_KEY);
            sessionId = null;
            var res = await apiFetch('/session', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json; charset=utf-8' },
                body: JSON.stringify({ domain_id: domainId })
            });
            var data = parseApiResponseJson(await res.text());
            if (!res.ok || !data.session_id) {
                throw new Error(data.error || 'Failed to create session');
            }
            sessionId = data.session_id;
            if (data.domain_id) domainId = data.domain_id;
            sessionStorage.setItem(STORAGE_KEY, sessionId);
            renderMessages([]);
            loadOnboarding(domainId);
        }

        el.domainSelect.addEventListener('change', function() {
            var next = el.domainSelect.value;
            if (next === domainId && sessionId) return;
            createSessionWithDomain(next).catch(function(e) {
                showToast(e.message || 'Domain switch error');
                el.domainSelect.value = domainId;
            });
        });

        function scrollToBottom() {
            requestAnimationFrame(function() {
                el.chatScroll.scrollTop = el.chatScroll.scrollHeight;
            });
        }

        /** Фото с сервера: fetch с initData → blob URL (тег img не шлёт auth сам). */
        async function loadAuthedImage(imgEl, imagePath) {
            try {
                var path = String(imagePath || '').replace(/^\/api\//, '');
                if (path.charAt(0) === '/') path = path.slice(1);
                var res = await apiFetch(path, { method: 'GET' });
                if (!res.ok) return;
                var blob = await res.blob();
                imgEl.src = URL.createObjectURL(blob);
            } catch (e) {
                console.error('loadAuthedImage', e);
            }
        }

        function renderMessages(messages) {
            el.messagesRoot.innerHTML = '';
            if (!messages || !messages.length) {
                var hint = document.createElement('div');
                hint.className = 'day-divider';
                hint.textContent = 'Напишите вопрос по базе знаний выбранного домена.';
                el.messagesRoot.appendChild(hint);
                updateOnboardingVisibility();
                return;
            }
            messages.forEach(function(m) {
                var row = document.createElement('div');
                row.className = 'row ' + (m.role === 'user' ? 'user' : 'assistant');
                var bubble = document.createElement('div');
                bubble.className = 'bubble';

                if (m.image_data_url || m.image_url) {
                    var img = document.createElement('img');
                    img.className = 'attach-preview';
                    img.alt = 'Фото пользователя';
                    if (m.image_data_url) {
                        img.src = m.image_data_url;
                    } else {
                        img.src = 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="120" height="80"><rect fill="#ddd" width="100%" height="100%"/></svg>');
                        loadAuthedImage(img, m.image_url);
                    }
                    bubble.appendChild(img);
                }
                if (m.content && String(m.content).trim()) {
                    var body = document.createElement('div');
                    body.className = 'body';
                    body.textContent = m.content;
                    bubble.appendChild(body);
                }
                if (m.role === 'user' && m.class_prediction) {
                    var meta = document.createElement('div');
                    meta.className = 'meta-line';
                    var pct = m.class_confidence > 0 ? Math.round(Number(m.class_confidence) * 100) : null;
                    meta.textContent = (m.class_prediction || '').replace(/_/g, ' ') + (pct != null ? ' · ' + pct + '%' : '');
                    bubble.appendChild(meta);
                }

                if (m.role === 'assistant' && m.citations && m.citations.length) {
                    var cites = document.createElement('div');
                    cites.className = 'citations';
                    var citesTitle = document.createElement('div');
                    citesTitle.className = 'citations-title';
                    citesTitle.textContent = 'Источники';
                    cites.appendChild(citesTitle);
                    m.citations.forEach(function(c, idx) {
                        var item = document.createElement('div');
                        item.className = 'citation-item';
                        var head = document.createElement('div');
                        head.className = 'citation-head';
                        head.textContent = '[' + (idx + 1) + '] ' + (c.filename || 'документ') +
                            (c.page ? ' · стр. ' + c.page : '');
                        item.appendChild(head);
                        if (c.excerpt) {
                            var ex = document.createElement('div');
                            ex.className = 'citation-excerpt';
                            ex.textContent = c.excerpt;
                            item.appendChild(ex);
                        }
                        cites.appendChild(item);
                    });
                    bubble.appendChild(cites);
                }

                if (m.role === 'assistant' && m.id) {
                    var fb = document.createElement('div');
                    fb.className = 'feedback-row';
                    var rated = m.feedback_rating;
                    [1, -1].forEach(function(r) {
                        var b = document.createElement('button');
                        b.type = 'button';
                        b.className = 'feedback-btn' + (rated === r ? ' active' : '');
                        b.setAttribute('data-rating', String(r));
                        b.setAttribute('data-feedback-for', String(m.id));
                        b.textContent = r === 1 ? '👍' : '👎';
                        b.disabled = rated != null;
                        b.addEventListener('click', function() { sendFeedback(m.id, r); });
                        fb.appendChild(b);
                    });
                    bubble.appendChild(fb);
                }

                row.appendChild(bubble);
                el.messagesRoot.appendChild(row);
            });
            updateOnboardingVisibility();
            scrollToBottom();
        }

        async function ensureSession() {
            var sid = sessionStorage.getItem(STORAGE_KEY);
            if (sid) {
                var hr = await apiFetch('/history?session_id=' + encodeURIComponent(sid), { method: 'GET' });
                if (hr.status === 404) {
                    sessionStorage.removeItem(STORAGE_KEY);
                    sid = null;
                } else if (hr.ok) {
                    var hd = parseApiResponseJson(await hr.text());
                    sessionId = hd.session_id || sid;
                    if (hd.domain_id) {
                        domainId = hd.domain_id;
                        sessionStorage.setItem(DOMAIN_STORAGE_KEY, domainId);
                        el.domainSelect.value = domainId;
                    }
                    renderMessages(hd.messages || []);
                    loadOnboarding(domainId);
                    return;
                } else {
                    sid = null;
                    sessionStorage.removeItem(STORAGE_KEY);
                }
            }
            var res = await apiFetch('/session', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json; charset=utf-8' },
                body: JSON.stringify({ domain_id: domainId })
            });
            var data = parseApiResponseJson(await res.text());
            if (!res.ok || !data.session_id) {
                throw new Error(data.error || 'Failed to create session');
            }
            sessionId = data.session_id;
            if (data.domain_id) {
                domainId = data.domain_id;
                sessionStorage.setItem(DOMAIN_STORAGE_KEY, domainId);
                el.domainSelect.value = domainId;
            }
            sessionStorage.setItem(STORAGE_KEY, sessionId);
            renderMessages([]);
            loadOnboarding(domainId);
        }

        function setSending(on) {
            sending = on;
            el.sendBtn.disabled = on;
            el.inputText.disabled = on;
            el.typingLine.classList.toggle('active', on);
        }

        async function sendMessageStream(text) {
            var url = '/message?stream=1';
            var res = await apiFetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json; charset=utf-8', 'Accept': 'text/event-stream' },
                body: JSON.stringify({ session_id: sessionId, domain_id: domainId, text: text })
            });
            if (!res.ok || !res.body) {
                throw new Error('Stream unavailable');
            }
            var reader = res.body.getReader();
            var decoder = new TextDecoder();
            var buffer = '';
            var draftRow = null;
            var draftBody = null;

            function ensureDraftBubble() {
                if (draftRow) return;
                draftRow = document.createElement('div');
                draftRow.className = 'row assistant';
                var bubble = document.createElement('div');
                bubble.className = 'bubble';
                draftBody = document.createElement('div');
                draftBody.className = 'body';
                bubble.appendChild(draftBody);
                draftRow.appendChild(bubble);
                el.messagesRoot.appendChild(draftRow);
                scrollToBottom();
            }

            while (true) {
                var chunk = await reader.read();
                if (chunk.done) break;
                buffer += decoder.decode(chunk.value, { stream: true });
                var parts = buffer.split('\n\n');
                buffer = parts.pop() || '';
                parts.forEach(function(block) {
                    var event = 'message';
                    var data = '';
                    block.split('\n').forEach(function(line) {
                        if (line.indexOf('event: ') === 0) event = line.slice(7).trim();
                        if (line.indexOf('data: ') === 0) data = line.slice(6);
                    });
                    if (!data) return;
                    if (event === 'token') {
                        try {
                            var tok = JSON.parse(data);
                            ensureDraftBubble();
                            draftBody.textContent += tok.text || '';
                            scrollToBottom();
                        } catch (e) { /* ignore */ }
                    } else if (event === 'done') {
                        var payload = JSON.parse(data);
                        if (payload.session_id) {
                            sessionId = payload.session_id;
                            sessionStorage.setItem(STORAGE_KEY, sessionId);
                        }
                        if (payload.messages) renderMessages(payload.messages);
                    } else if (event === 'error') {
                        var err = JSON.parse(data);
                        throw new Error(err.error || 'Stream error');
                    }
                });
            }
        }

        async function sendMessage() {
            if (sending) return;
            var text = (el.inputText.value || '').trim();
            if (!text) {
                showToast('Enter a message');
                return;
            }
            if (!sessionId) {
                try { await ensureSession(); } catch (e) {
                    showToast(e.message || 'Session error');
                    return;
                }
            }

            setSending(true);
            try {
                if (window.ReadableStream && typeof TextDecoder !== 'undefined') {
                    try {
                        await sendMessageStream(text);
                        el.inputText.value = '';
                        autoResize();
                        return;
                    } catch (streamErr) {
                        console.warn('stream fallback', streamErr);
                    }
                }
                var res = await apiFetch('/message', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json; charset=utf-8' },
                    body: JSON.stringify({ session_id: sessionId, domain_id: domainId, text: text })
                });
                var data = parseApiResponseJson(await res.text());
                if (data.session_id) {
                    sessionId = data.session_id;
                    sessionStorage.setItem(STORAGE_KEY, sessionId);
                }
                if (data.domain_id) {
                    domainId = data.domain_id;
                    sessionStorage.setItem(DOMAIN_STORAGE_KEY, domainId);
                    el.domainSelect.value = domainId;
                }
                if (data.messages) {
                    renderMessages(data.messages);
                }
                if (!res.ok) {
                    showToast(data.error || ('Error ' + res.status));
                } else if (data.error) {
                    showToast(data.error);
                }
                el.inputText.value = '';
                autoResize();
            } catch (e) {
                console.error(e);
                showToast(e.message || 'Network error');
            } finally {
                setSending(false);
            }
        }

        function autoResize() {
            var ta = el.inputText;
            ta.style.height = 'auto';
            ta.style.height = Math.min(ta.scrollHeight, 120) + 'px';
        }

        el.sendBtn.addEventListener('click', sendMessage);
        el.inputText.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage();
            }
        });
        el.inputText.addEventListener('input', autoResize);

        loadBranding().then(function() {
            return loadDomainsCatalog();
        }).then(function() {
            return ensureSession();
        }).then(function() {
            return loadOnboarding(domainId);
        }).catch(function(e) {
            console.error(e);
            showToast(e.message || 'Connection failed');
        });
