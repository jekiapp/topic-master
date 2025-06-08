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
        $tailContent.empty();
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
        var limitMsg = parseInt($tailLimitMsg.val(), 3);
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
                        var msgHtml = '<div class="tail-msg"><span class="tail-topic">[' + obj.topic + ']</span> <span class="tail-payload">' + escapeHtml(obj.payload) + '</span></div>';
                        $tailContent.append(msgHtml);
                        $tailContent.scrollTop($tailContent[0].scrollHeight);
                    } catch (e) {
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
