// Tail panel logic for topic details

function initTailPanel({getCurrentTopicDetail, adjustPanelWidths}) {
    var tailSocket = null;
    var $tailPanel = $('#tail-panel');
    var $tailBtn = $('.btn-tail');
    var $tailPanelBtn = $('#tail-panel-btn.btn-tail-panel');
    var $tailStopBtn = $('#tail-panel-btn.btn-stop-tail-panel');
    var $tailCloseBtn = $('#close-tail-panel');
    var $tailContent = $('#tail-content');
    var $tailStatus = $('#tail-status');
    var $tailLimitMsg = $('#tail-limit-msg');

    function setTailingActive(active) {
        if (active) {
            $tailPanelBtn.prop('disabled', true);
            $tailStopBtn.prop('disabled', false).show();
        } else {
            $tailPanelBtn.prop('disabled', false);
            $tailStopBtn.prop('disabled', true).hide();
        }
    }
    setTailingActive(false);

    $tailBtn.on('click', function() {
        $tailPanel.show();
        $tailBtn.prop('disabled', true);
        $tailPanelBtn.prop('disabled', false);
        $tailStopBtn.prop('disabled', true).hide();
        $tailStatus.text('');
        if (adjustPanelWidths) adjustPanelWidths();
    });
    $tailCloseBtn.on('click', function() {
        $tailPanel.hide();
        $tailBtn.prop('disabled', false);
        if (tailSocket) {
            tailSocket.close();
            tailSocket = null;
        }
        setTailingActive(false);
        $tailStatus.text('');
        $tailContent.empty();
        if (adjustPanelWidths) adjustPanelWidths();
    });
    $tailPanelBtn.on('click', function() {
        var currentTopicDetail = getCurrentTopicDetail();
        if (!currentTopicDetail) {
            $tailStatus.text('Topic detail not loaded').css('color', 'red');
            return;
        }
        var limitMsg = parseInt($tailLimitMsg.val(), 10);
        if (!limitMsg || limitMsg <= 0) {
            $tailStatus.text('Limit must be > 0').css('color', 'red');
            return;
        }
        setTailingActive(true);
        $tailStatus.text('Connecting...').css('color', '#888');
        $tailContent.empty();
        if (tailSocket) {
            tailSocket.close();
            tailSocket = null;
        }
        // Build query string for WebSocket URL
        var topic = encodeURIComponent(currentTopicDetail.name);
        var limitMsgStr = encodeURIComponent(limitMsg);
        var hosts = (currentTopicDetail.nsqd_hosts || []).map(encodeURIComponent);
        var params = `topic=${topic}&limit_msg=${limitMsgStr}`;
        hosts.forEach(function(h) { params += `&nsqd_hosts=${h}`; });
        var wsProto = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        var wsUrl = wsProto + window.location.host + '/api/topic/tail?' + params;
        tailSocket = new WebSocket(wsUrl);
        tailSocket.onopen = function() {
            $tailStatus.text('Connected. Waiting for messages...').css('color', '#888');
        };
        tailSocket.onmessage = function(event) {
            // Split by record separator (ASCII 30)
            var RS = String.fromCharCode(30);
            var parts = event.data.split(RS);
            parts.forEach(function(part) {
                if (part.trim()) {
                    try {
                        var obj = JSON.parse(part);
                        var timestamp = '<span class="tail-timestamp">[' + obj.timestamp + ']</span>';
                        var body = '<span class="tail-body">' + escapeHtml(obj.payload) + '</span>';
                        var prettyBtn = '<span class="tail-pretty-btn" title="Pretty print JSON" style="cursor:pointer;user-select:none;margin-left:8px;font-size:1.1em;">✨</span>';
                        var copyBtn = '<span class="tail-copy-btn" title="Copy to clipboard" style="cursor:pointer;user-select:none;margin-left:8px;font-size:1.1em;">📄</span>';
                        var msgHtml = '<div class="tail-msg">' + timestamp + body + copyBtn + prettyBtn + '</div>';

                        $tailContent.prepend(msgHtml);

                        var $msg = $tailContent.find('.tail-msg').first();
                        $msg.find('.tail-copy-btn').off('click').on('click', function() {
                            navigator.clipboard.writeText(obj.payload);
                            var $btn = $(this);
                            if ($btn.next('.tail-copied-label').length === 0) {
                                var $label = $('<span class="tail-copied-label" style="margin-left:4px;color:#2ecc40;font-size:0.98em;">Copied!</span>');
                                $btn.after($label);
                                setTimeout(function() { $label.fadeOut(200, function() { $label.remove(); }); }, 1200);
                            }
                        });
                        $msg.find('.tail-pretty-btn').off('click').on('click', function() {
                            var $body = $msg.find('.tail-body');
                            var raw = $body.data('raw');
                            if (raw === undefined) {
                                raw = $body.text();
                                $body.data('raw', raw);
                            }
                            if ($body.data('pretty')) {
                                $body.text(raw);
                                $body.data('pretty', false);
                            } else {
                                try {
                                    var parsed = JSON.parse(raw);
                                    var pretty = JSON.stringify(parsed, null, 2);
                                    $body.html('<pre style="margin:0;font-family:monospace;font-size:0.98em;">' + escapeHtml(pretty) + '</pre>');
                                    $body.data('pretty', true);
                                } catch (e) {
                                    $body.text(raw);
                                    $body.data('pretty', false);
                                }
                            }
                        });

                    } catch (e) {
                        console.error(e);
                        $tailContent.append('<div class="tail-msg tail-msg-error">' + escapeHtml(part) + '</div>');
                    }
                }
            });
        };
        tailSocket.onerror = function() {
            $tailStatus.text('WebSocket error').css('color', 'red');
            setTailingActive(false);
        };
        tailSocket.onclose = function() {
            $tailStatus.text('Connection closed').css('color', '#888');
            setTailingActive(false);
            $tailBtn.prop('disabled', false);
        };
    });
    $tailStopBtn.on('click', function() {
        if (tailSocket) {
            tailSocket.close();
            tailSocket = null;
            $tailStatus.text('Stopped by user.').css('color', '#888');
        }
        setTailingActive(false);
        $tailBtn.prop('disabled', false);
    });

    // Helper to escape HTML
    function escapeHtml(text) {
        return text.replace(/[&<>"]'/g, function(m) {
            return ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#39;'}[m]);
        });
    }
}

window.initTailPanel = initTailPanel;
