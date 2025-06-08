$(function() {
    var topicID = getTopicNameFromURL();
    if (!topicID) {
        alert('Topic name is required');
        return;
    }

    // --- Cache for topic detail ---
    var currentTopicDetail = null;

    // Fetch topic detail ONCE
    $.get('/api/topic/detail', { topic: topicID }, function(response) {
        if (!response || !response.data) {
            alert('Failed to load topic detail');
            return;
        }
        var detail = response.data;
        currentTopicDetail = detail;
        $('.topic-name').text(detail.name);
        $('.group-owner').text(detail.group_owner);
        var $eventTrigger = $('.event-trigger-input');
        $eventTrigger.val(detail.event_trigger);
        $eventTrigger.prop('readonly', !detail.permission.can_update_event_trigger);
        $eventTrigger.data('original', detail.event_trigger);

        // Bookmark icon
        var bookmarkImg = $('.bookmark-img');
        if (detail.bookmarked) {
            bookmarkImg.attr('src', '/icons/bookmark-true.png');
        } else {
            bookmarkImg.attr('src', '/icons/bookmark-false.png');
        }

        // Render nsqd hosts
        var hostsList = $('.nsqd-hosts-list');
        hostsList.empty();
        $.each(detail.nsqd_hosts, function(_, host) {
            hostsList.append($('<li>').text(host));
        });

        // Enable/disable buttons based on permission
        $('.btn-pause').prop('disabled', !detail.permission.can_pause);
        $('.btn-publish').prop('disabled', !detail.permission.can_publish);
        $('.btn-tail').prop('disabled', !detail.permission.can_tail);
        $('.btn-delete').prop('disabled', !detail.permission.can_delete);
        $('.btn-drain').prop('disabled', !detail.permission.can_empty_queue);

        // Event trigger checkmark logic
        var $check = $('.event-trigger-check');
        $eventTrigger.on('input', function() {
            var orig = $eventTrigger.data('original');
            if ($eventTrigger.val() !== orig) {
                $check.show();
            } else {
                $check.hide();
            }
        });

        // Fetch topic stats using hosts and topic name
        fetchAndUpdateStats(detail);
    }).fail(function() {
        alert('Failed to load topic detail');
    });

    // Back button (optional: history.back or custom logic)
    $('#back-link').on('click', function() {
        window.history.back();
    });

    // --- Autorefresh logic ---
    var autorefreshCountdown = 0;
    var autorefreshActive = false;
    var autorefreshInterval = null;
    var $autorefresh = $('#autorefresh-link');
    var $autorefreshIcon = $('#autorefresh-icon');
    var originalIconHtml = $autorefresh.html();

    function fetchAndUpdateStats(detail) {
        var hostsStr = (detail.nsqd_hosts || []).join(',');
        $.get('/api/topic/stats', { hosts: hostsStr, topic: detail.name }, function(statsResp) {
            if (!statsResp || !statsResp.data) {
                $('.topic-stats-depth').text('-');
                $('.topic-stats-messages').text('-');
                return;
            }
            $('.topic-stats-depth').text(statsResp.data.depth);
            $('.topic-stats-messages').text(statsResp.data.messages);
        }).fail(function() {
            $('.topic-stats-depth').text('-');
            $('.topic-stats-messages').text('-');
        });
    }

    // Only refresh stats, not topic detail
    function refreshStats() {
        if (!currentTopicDetail) return;
        fetchAndUpdateStats(currentTopicDetail);
    }

    $autorefresh.on('click', function() {
        if (autorefreshActive) return;
        autorefreshActive = true;
        autorefreshCountdown = 30;
        $autorefresh.html('Refreshing...' + autorefreshCountdown);
        $autorefresh.css('pointer-events', 'none');
        refreshStats();
        autorefreshInterval = setInterval(function() {
            autorefreshCountdown--;
            if (autorefreshCountdown > 0) {
                $autorefresh.html('Refreshing...' + autorefreshCountdown);
                if (autorefreshCountdown % 2 === 0) {
                    refreshStats();
                }
            } else {
                clearInterval(autorefreshInterval);
                $autorefresh.html(originalIconHtml);
                $autorefresh.css('pointer-events', 'auto');
                autorefreshActive = false;
            }
        }, 1000);
    });

    // --- Publish panel logic ---
    $('.btn-publish').on('click', function() {
        $('#publish-panel').show();
        $('.btn-publish').prop('disabled', true);
    });
    $('#close-publish-panel').on('click', function() {
        $('#publish-panel').hide();
        $('#publish-textarea').val('');
        $('#publish-status').text('').css('color', '');
        $('.btn-publish').prop('disabled', false);
    });
    // Publish button handler
    $('#publish-panel-btn').on('click', function() {
        var message = $('#publish-textarea').val();
        var $status = $('#publish-status');
        $status.text('').css('color', '');
        if (!message) {
            $status.text('Message cannot be empty').css('color', 'red');
            return;
        }
        if (!currentTopicDetail) {
            $status.text('Topic detail not loaded').css('color', 'red');
            return;
        }
        var payload = {
            topic: currentTopicDetail.name,
            message: message,
            nsqd_hosts: currentTopicDetail.nsqd_hosts
        };
        $.ajax({
            url: '/api/topic/publish',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(payload),
            success: function(resp) {
                $status.text('Message published').css('color', 'green');
                $('.btn-publish').prop('disabled', false);
            },
            error: function(xhr) {
                var msg = 'Failed to publish';
                if (xhr.responseJSON && xhr.responseJSON.error) {
                    msg += ': ' + xhr.responseJSON.error;
                }
                $status.text(msg).css('color', 'red');
            }
        });
    });

    // --- Tail panel logic ---
    var tailSocket = null;
    var $tailPanel = $('#tail-panel');
    var $tailBtn = $('.btn-tail');
    var $tailPanelBtn = $('#tail-panel-btn');
    var $tailCloseBtn = $('#close-tail-panel');
    var $tailContent = $('#tail-content');
    var $tailStatus = $('#tail-status');
    var $tailLimitMsg = $('#tail-limit-msg');

    $tailBtn.on('click', function() {
        $tailPanel.show();
        $tailBtn.prop('disabled', true);
        $tailPanelBtn.prop('disabled', false);
        $tailContent.empty();
        $tailStatus.text('');
    });
    $tailCloseBtn.on('click', function() {
        $tailPanel.hide();
        $tailBtn.prop('disabled', false);
        if (tailSocket) {
            tailSocket.close();
            tailSocket = null;
        }
        $tailPanelBtn.prop('disabled', false);
        $tailStatus.text('');
        $tailContent.empty();
    });
    $tailPanelBtn.on('click', function() {
        if (!currentTopicDetail) {
            $tailStatus.text('Topic detail not loaded').css('color', 'red');
            return;
        }
        var limitMsg = parseInt($tailLimitMsg.val(), 10);
        if (!limitMsg || limitMsg <= 0) {
            $tailStatus.text('Limit must be > 0').css('color', 'red');
            return;
        }
        $tailPanelBtn.prop('disabled', true);
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
            $tailPanelBtn.prop('disabled', false);
        };
        tailSocket.onclose = function() {
            $tailStatus.text('Connection closed').css('color', '#888');
            $tailPanelBtn.prop('disabled', false);
            $tailBtn.prop('disabled', false);
        };
    });

    // Helper to escape HTML
    function escapeHtml(text) {
        return text.replace(/[&<>"']/g, function(m) {
            return ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#39;'}[m]);
        });
    }
});

// Helper: get topic name from URL (e.g., ?topic=MyTopic)
function getTopicNameFromURL() {
    var params = new URLSearchParams(window.location.search);
    return params.get('id');
} 