// --- Local login helpers for iframe context ---
function isLogin() {
    return !!localStorage.getItem('user');
}
function getUserInfo() {
    var user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
}

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
                if (!window.parent.isLogin || !window.parent.isLogin()) {
                    alert('Please log in to bookmark topics.');
                    return;
                }
                $.ajax({
                    url: '/api/entity/toggle-bookmark',
                    method: 'POST',
                    contentType: 'application/json',
                    data: JSON.stringify({ entity_id: detail.id, bookmark: !detail.bookmarked }),
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
        if (window.parent.isLogin && window.parent.isLogin()) {
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

    // --- Action buttons: Pause, Delete, Empty Queue ---
    $('.btn-pause').on('click', function() {
        if (!currentTopicDetail) return;
        var $btn = $(this);
        $btn.prop('disabled', true);
        $.ajax({
            url: '/api/topic/nsq/pause',
            method: 'GET',
            contentType: 'application/json',
            data: JSON.stringify({ id: currentTopicDetail.id }),
            success: function(resp) {
                showStatus('Topic paused successfully', 'green');
            },
            error: function(xhr) {
                var msg = 'Failed to pause topic';
                if (xhr.responseJSON && xhr.responseJSON.error) {
                    msg += ': ' + xhr.responseJSON.error;
                }
                showStatus(msg, 'red');
            },
            complete: function() {
                $btn.prop('disabled', false);
            }
        });
    });

    $('.btn-delete').on('click', function() {
        if (!currentTopicDetail) return;
        var modalHtml = [
            '<div style="text-align:center;">',
            '<div style="font-size:1.1em;margin-bottom:18px;">Are you sure you want to delete this topic? This cannot be undone.</div>',
            '<button id="modal-delete-confirm" style="margin-right:18px;padding:8px 18px;background:#ff2d2d;color:#fff;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Yes, Delete</button>',
            '<button id="modal-delete-cancel" style="padding:8px 18px;background:#eee;color:#333;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Cancel</button>',
            '</div>'
        ].join('');
        window.parent.showModalOverlay(modalHtml);
        setTimeout(function() {
            $('#modal-delete-confirm', window.parent.document).off('click').on('click', function() {
                var $btn = $('.btn-delete');
                $btn.prop('disabled', true);
                window.parent.hideModalOverlay();
                $.ajax({
                    url: '/api/topic/delete',
                    method: 'GET',
                    contentType: 'application/json',
                    data: JSON.stringify({ id: currentTopicDetail.id }),
                    success: function(resp) {
                        showStatus('Topic deleted successfully', 'green');
                        setTimeout(function() { window.location.href = '/'; }, 1200);
                    },
                    error: function(xhr) {
                        var msg = 'Failed to delete topic';
                        if (xhr.responseJSON && xhr.responseJSON.error) {
                            msg += ': ' + xhr.responseJSON.error;
                        }
                        showStatus(msg, 'red');
                    },
                    complete: function() {
                        $btn.prop('disabled', false);
                    }
                });
            });
            $('#modal-delete-cancel', window.parent.document).off('click').on('click', function() {
                window.parent.hideModalOverlay();
            });
        }, 0);
    });

    $('.btn-empty').on('click', function() {
        if (!currentTopicDetail) return;
        // Use modal overlay for confirmation
        
            var modalHtml = [
                '<div style="text-align:center;">',
                '<div style="font-size:1.1em;margin-bottom:18px;">Are you sure you want to empty the queue for this topic? This cannot be undone.</div>',
                '<button id="modal-empty-confirm" style="margin-right:18px;padding:8px 18px;background:#ff2d2d;color:#fff;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Yes, Empty</button>',
                '<button id="modal-empty-cancel" style="padding:8px 18px;background:#eee;color:#333;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Cancel</button>',
                '</div>'
            ].join('');
            window.parent.showModalOverlay(modalHtml);
            // Attach handlers after modal is shown
            setTimeout(function() {
                $('#modal-empty-confirm', window.parent.document).off('click').on('click', function() {
                    var $btn = $('.btn-empty');
                    $btn.prop('disabled', true);
                    window.parent.hideModalOverlay();
                    $.ajax({
                        url: '/api/topic/nsq/empty',
                        method: 'GET',
                        contentType: 'application/json',
                        data: JSON.stringify({ id: currentTopicDetail.id }),
                        success: function(resp) {
                            showStatus('Queue emptied successfully', 'green');
                            refreshStats();
                        },
                        error: function(xhr) {
                            var msg = 'Failed to empty queue';
                            if (xhr.responseJSON && xhr.responseJSON.error) {
                                msg += ': ' + xhr.responseJSON.error;
                            }
                            showStatus(msg, 'red');
                        },
                        complete: function() {
                            $btn.prop('disabled', false);
                        }
                    });
                });
                $('#modal-empty-cancel', window.parent.document).off('click').on('click', function() {
                    window.parent.hideModalOverlay();
                });
        }, 0);
    });

    // Helper to show status messages
    function showStatus(msg, color) {
        var $status = $('#topic-action-status');
        if (!$status.length) {
            $('.topic-actions').append('<div id="topic-action-status" style="margin-top:6px;font-size:0.98em;"></div>');
            $status = $('#topic-action-status');
        }
        $status.text(msg).css('color', color || 'black');
        if (color === 'green') {
            setTimeout(function() { $status.text(''); }, 2000);
        }
    }
});

// Helper: get topic name from URL (e.g., ?topic=MyTopic)
function getTopicNameFromURL() {
    var params = new URLSearchParams(window.location.search);
    return params.get('id');
} 