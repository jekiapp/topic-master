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
        var $check = $('.event-trigger-check');
        var $reset = $('.event-trigger-reset');

        function updateEventTriggerButtons() {
            var orig = $eventTrigger.data('original');
            if ($eventTrigger.val() !== orig) {
                $check.show();
                $reset.show();
            } else {
                $check.hide();
                $reset.hide();
            }
        }
        $eventTrigger.on('input', function() {
            updateEventTriggerButtons();
        });
        $check.on('click', function() {
            var newValue = $eventTrigger.val();
            $.ajax({
                url: '/api/entity/update-description',
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ entity_id: detail.id, description: newValue }),
                success: function() {
                    $eventTrigger.data('original', newValue);
                    updateEventTriggerButtons();
                    // $check.fadeOut(120).fadeIn(120);
                },
                error: function() {
                    alert('Failed to update event trigger');
                }
            });
        });
        $reset.on('click', function() {
            $eventTrigger.val($eventTrigger.data('original'));
            updateEventTriggerButtons();
        });
        updateEventTriggerButtons();

        // Bookmark icon
        var bookmarkImg = $('.bookmark-img');
        function setBookmarkIcon(state) {
            if (state) {
                bookmarkImg.attr('src', '/icons/bookmark-true.png');
            } else {
                bookmarkImg.attr('src', '/icons/bookmark-false.png');
            }
        }
        setBookmarkIcon(detail.bookmarked);

        // --- Bookmark toggle logic ---
        function enableBookmarkToggle() {
            bookmarkImg.css('cursor', 'pointer');
            bookmarkImg.attr('title', detail.bookmarked ? 'Remove Bookmark' : 'Add Bookmark');
            bookmarkImg.off('click').on('click', function(e) {
                e.preventDefault();
                if (!window.isLogin || !window.isLogin()) {
                    alert('Please log in to bookmark topics.');
                    return;
                }
                $.ajax({
                    url: '/api/bookmark/toggle',
                    method: 'POST',
                    contentType: 'application/json',
                    data: JSON.stringify({ entity_id: detail.name }),
                    success: function(resp) {
                        // Toggle state locally for instant feedback
                        detail.bookmarked = !detail.bookmarked;
                        setBookmarkIcon(detail.bookmarked);
                        bookmarkImg.attr('title', detail.bookmarked ? 'Remove Bookmark' : 'Add Bookmark');
                    },
                    error: function(xhr) {
                        var msg = 'Failed to toggle bookmark';
                        if (xhr.responseJSON && xhr.responseJSON.error) {
                            msg += ': ' + xhr.responseJSON.error;
                        }
                        alert(msg);
                    }
                });
            });
        }
        function disableBookmarkToggle() {
            bookmarkImg.css('cursor', 'not-allowed');
            bookmarkImg.attr('title', 'Log in to bookmark');
            bookmarkImg.off('click');
        }
        if (window.isLogin && window.isLogin()) {
            enableBookmarkToggle();
        } else {
            disableBookmarkToggle();
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
    function adjustPanelWidths() {
        var publishVisible = $('#publish-panel').is(':visible');
        var tailVisible = $('#tail-panel').is(':visible');
        if (publishVisible && !tailVisible) {
            $('#publish-panel').addClass('wide-panel');
        } else {
            $('#publish-panel').removeClass('wide-panel');
        }
        if (tailVisible && !publishVisible) {
            $('#tail-panel').addClass('wide-panel');
        } else {
            $('#tail-panel').removeClass('wide-panel');
        }
    }

    $('.btn-publish').on('click', function() {
        $('#publish-panel').show();
        $('.btn-publish').prop('disabled', true);
        adjustPanelWidths();
    });
    $('#close-publish-panel').on('click', function() {
        $('#publish-panel').hide();
        $('#publish-textarea').val('');
        $('#publish-status').text('').css('color', '');
        $('.btn-publish').prop('disabled', false);
        adjustPanelWidths();
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

    // --- Tail panel logic moved to tail_msg.js ---
    if (window.initTailPanel) {
        window.initTailPanel({
            getCurrentTopicDetail: function() { return currentTopicDetail; },
            adjustPanelWidths: adjustPanelWidths
        });
    }

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