// --- Local login helpers for iframe context ---
function isLogin() {
    return !!localStorage.getItem('user');
}
function getUserInfo() {
    var user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
}

$(function() {
    var topicID = getTopicIDFromURL();
    if (!topicID) {
        alert('Topic name is required');
        return;
    }

    // --- Cache for topic detail ---
    var currentTopicDetail = null;
    window.currentTopicDetail = null;

    // Fetch topic detail ONCE
    $.get('/api/topic/detail', { topic: topicID }, function(response) {
        if (!response || !response.data) {
            alert('Failed to load topic detail');
            return;
        }
        var detail = response.data;
        currentTopicDetail = detail;
        window.currentTopicDetail = detail;
        $('.topic-name').text(detail.name);
        $('.group-owner').text(detail.group_owner);
        var $eventTrigger = $('.event-trigger-input');
        $eventTrigger.val(detail.event_trigger);
        $eventTrigger.prop('readonly', !detail.is_free_action);
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
            bookmarkImg.attr('title', detail.bookmarked ? 'Remove Bookmark' : 'Add Bookmark');
            bookmarkImg.off('click').on('click', function(e) {
                e.preventDefault();
                if (!window.parent.isLogin || !window.parent.isLogin()) {
                    if (window.parent && window.parent.showModalOverlay) {
                        window.parent.showModalOverlay('You need to login to bookmark topic');
                    } else {
                        alert('You need to login to bookmark topic');
                    }
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
            bookmarkImg.attr('title', 'Log in to bookmark');
            bookmarkImg.off('click').on('click', function(e) {
                e.preventDefault();
                if (window.parent && window.parent.showModalOverlay) {
                    window.parent.showModalOverlay('You need to login to bookmark topic');
                } else {
                    alert('You need to login to bookmark topic');
                }
            });
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
            if (typeof host === 'object' && host.host_name && host.address) {
                hostsList.append(
                    $('<li>').html(
                        escapeHtml(host.host_name) +
                        ' <span style="font-size:0.92em;color:#888;">(' + escapeHtml(host.address) + ')</span>'
                    )
                );
            } else {
                // fallback for old string format
                hostsList.append($('<li>').text(host));
            }
        });

        // Helper for permission and login check (async)
        function checkActionPermissionAsync(isFreeAction, groupOwner, actionName, entityId, cb) {
            if (isFreeAction) { cb(true); return; }
            if (!(window.parent.isLogin && window.parent.isLogin())) {
                window.parent.showModalOverlay(`This topic is owned by ${escapeHtml(groupOwner)}. You must login to perform this action`);
                cb(false); return;
            }
            // User is logged in, check with backend
            $.ajax({
                url: '/api/auth/check-action',
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ action: actionName, entity_id: entityId }),
                success: function(resp) {
                    resp = resp.data;
                    if (resp.allowed) {
                        cb(true);
                    } else {
                        urlApply = "#tickets-new?type=topic_action&entity_id=" + entityId + "&action=" + actionName;
                        window.parent.showModalOverlay(`You do not have permission to perform this action. <br/><br/><a href="${urlApply}" target="_blank">Apply for permission</a>`);
                        cb(false);
                    }
                },
                error: function(xhr) {
                    let msg = 'Permission check failed';
                    if (xhr.responseJSON && xhr.responseJSON.error) msg += ': ' + xhr.responseJSON.error;
                    window.parent.showModalOverlay(msg);
                    cb(false);
                }
            });
        }

        // Toggle Pause/Resume button visibility based on paused status
        var $pauseBtn = $('.btn-pause');
        var $resumeBtn = $('.btn-resume');
        if (detail.platform_status && detail.platform_status.is_paused) {
            $pauseBtn.hide();
            $resumeBtn.show();
        } else {
            $pauseBtn.show();
            $resumeBtn.hide();
        }

        // update channel here
        refreshChannels(detail);

        // Claim link
        $('.claim-link').off('click').on('click', function() {
            window.showClaimModal(detail.id, detail.name, window.handleClaimEntity);
        });

        // --- Pause button logic ---
        $('.btn-pause').off('click').on('click', function() {
            if (!currentTopicDetail) return;
            checkActionPermissionAsync(
                currentTopicDetail.is_free_action,
                currentTopicDetail.group_owner,
                'pause',
                currentTopicDetail.id,
                function(allowed) {
                    if (!allowed) return;
                    var modalHtml = [
                        '<div style="text-align:center;">',
                        '<div style="font-size:1.1em;margin-bottom:18px;">Are you sure you want to pause this topic?</div>',
                        '<button id="modal-pause-confirm" style="margin-right:18px;padding:8px 18px;background:#ff2d2d;color:#fff;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Yes, Pause</button>',
                        '<button id="modal-pause-cancel" style="padding:8px 18px;background:#eee;color:#333;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Cancel</button>',
                        '</div>'
                    ].join('');
                    window.parent.showModalOverlay(modalHtml);
                    setTimeout(function() {
                        $('#modal-pause-confirm', window.parent.document).off('click').on('click', function() {
                            window.parent.hideModalOverlay();
                            var $btn = $('.btn-pause');
                            $btn.prop('disabled', true);
                            $.ajax({
                                url: '/api/topic/nsq/pause?id=' + currentTopicDetail.id + '&entity_id=' + encodeURIComponent(currentTopicDetail.id),
                                method: 'GET',
                                success: function(resp) {
                                    showStatus('Topic paused successfully', 'green');
                                    location.reload();
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
                        $('#modal-pause-cancel', window.parent.document).off('click').on('click', function() {
                            window.parent.hideModalOverlay();
                        });
                    }, 100);
                }
            );
        });

        // --- Resume button logic ---
        $('.btn-resume').off('click').on('click', function() {
            if (!currentTopicDetail) return;
            checkActionPermissionAsync(
                currentTopicDetail.is_free_action,
                currentTopicDetail.group_owner,
                'pause',
                currentTopicDetail.id,
                function(allowed) {
                    if (!allowed) return;
                    var modalHtml = [
                        '<div style="text-align:center;">',
                        '<div style="font-size:1.1em;margin-bottom:18px;">Are you sure you want to resume this topic?</div>',
                        '<button id="modal-resume-confirm" style="margin-right:18px;padding:8px 18px;background:#c7efc0;color:#333;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Yes, Resume</button>',
                        '<button id="modal-resume-cancel" style="padding:8px 18px;background:#eee;color:#333;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Cancel</button>',
                        '</div>'
                    ].join('');
                    window.parent.showModalOverlay(modalHtml);
                    setTimeout(function() {
                        $('#modal-resume-confirm', window.parent.document).off('click').on('click', function() {
                            window.parent.hideModalOverlay();
                            var $btn = $('.btn-resume');
                            $btn.prop('disabled', true);
                            $.ajax({
                                url: '/api/topic/nsq/resume?id=' + currentTopicDetail.id + '&entity_id=' + encodeURIComponent(currentTopicDetail.id),
                                method: 'GET',
                                success: function(resp) {
                                    showStatus('Topic resumed successfully', 'green');
                                    location.reload();
                                },
                                error: function(xhr) {
                                    var msg = 'Failed to resume topic';
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
                        $('#modal-resume-cancel', window.parent.document).off('click').on('click', function() {
                            window.parent.hideModalOverlay();
                        });
                    }, 100);
                }
            );
        });

        // --- Delete button logic ---
        $('.btn-delete').off('click').on('click', function() {
            if (!currentTopicDetail) return;
            checkActionPermissionAsync(
                currentTopicDetail.is_free_action,
                currentTopicDetail.group_owner,
                'delete',
                currentTopicDetail.id,
                function(allowed) {
                    if (!allowed) return;
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
                                url: '/api/topic/delete?id=' + currentTopicDetail.id + '&entity_id=' + encodeURIComponent(currentTopicDetail.id),
                                method: 'GET',
                                success: function(resp) {
                                    showStatus('Topic deleted successfully', 'green');
                                    setTimeout(function() {
                                        var hash = window.parent.location.hash || '';
                                        var backMatch = hash.match(/back=([^&]+)/);
                                        var back = backMatch ? decodeURIComponent(backMatch[1]) : null;
                                        if (back) {
                                            window.parent.location.hash = `#${back}`;
                                        } else {
                                            window.history.back();
                                        }
                                    }, 1200);
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
                }
            );
        });

        // --- Empty/Drain button logic ---
        $('.btn-drain, .btn-empty').off('click').on('click', function() {
            if (!currentTopicDetail) return;
            checkActionPermissionAsync(
                currentTopicDetail.is_free_action,
                currentTopicDetail.group_owner,
                'empty',
                currentTopicDetail.id,
                function(allowed) {
                    if (!allowed) return;
                    var modalHtml = [
                        '<div style="text-align:center;">',
                        '<div style="font-size:1.1em;margin-bottom:18px;">Are you sure you want to empty the queue for this topic? This cannot be undone.</div>',
                        '<button id="modal-empty-confirm" style="margin-right:18px;padding:8px 18px;background:#ff2d2d;color:#fff;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Yes, Empty</button>',
                        '<button id="modal-empty-cancel" style="padding:8px 18px;background:#eee;color:#333;border:none;border-radius:6px;font-weight:600;cursor:pointer;">Cancel</button>',
                        '</div>'
                    ].join('');
                    window.parent.showModalOverlay(modalHtml);
                    setTimeout(function() {
                        $('#modal-empty-confirm', window.parent.document).off('click').on('click', function() {
                            var $btn = $('.btn-empty');
                            $btn.prop('disabled', true);
                            window.parent.hideModalOverlay();
                            $.ajax({
                                url: '/api/topic/nsq/empty?id=' + currentTopicDetail.id + '&entity_id=' + encodeURIComponent(currentTopicDetail.id),
                                method: 'GET',
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
                }
            );
        });

        // --- Publish button logic ---
        $('.btn-publish').off('click').on('click', function() {
            if (!currentTopicDetail) return;
            checkActionPermissionAsync(
                currentTopicDetail.is_free_action,
                currentTopicDetail.group_owner,
                'publish',
                currentTopicDetail.id,
                function(allowed) {
                    if (!allowed) return;
                    $('#publish-panel').show();
                    $('.btn-publish').prop('disabled', true);
                    adjustPanelWidths();
                }
            );
        });
    }).fail(function() {
        alert('Failed to load topic detail');
    });

    // Back button (optional: history.back or custom logic)
    $('#back-link').on('click', function() {
        var hash = window.parent.location.hash || '';
        var backMatch = hash.match(/back=([^&]+)/);
        var back = backMatch ? decodeURIComponent(backMatch[1]) : null;
        if (back) {
            window.parent.location.hash = `#${back}`;
        } else {
            window.history.back();
        }
    });

    // --- Autorefresh logic ---
    var autorefreshCountdown = 0;
    var autorefreshActive = false;
    var autorefreshInterval = null;
    var $autorefresh = $('#autorefresh-link');
    var $autorefreshIcon = $('#autorefresh-icon');
    var originalIconHtml = $autorefresh.html();

    function fetchAndUpdateStats(detail) {
        var hostsArr = (detail.nsqd_hosts || []).map(function(host) {
            if (typeof host === 'object' && host.address) {
                return host.address;
            }
            return host;
        });
        var hostsStr = hostsArr.join(',');
        $.get('/api/topic/stats', { hosts: hostsStr, topic: detail.name }, function(statsResp) {
            if (!statsResp || !statsResp.data) {
                $('.topic-stats-depth').text('-');
                $('.topic-stats-messages').text('-');
                return;
            }
            $('.topic-stats-depth').text(Number(statsResp.data.depth).toLocaleString());
            $('.topic-stats-messages').text(Number(statsResp.data.messages).toLocaleString());

            // --- Update channel stats if present ---
            if (statsResp.data.channel_stats) {
                Object.entries(statsResp.data.channel_stats).forEach(function([channelName, stats]) {
                    var $row = $(`tr[data-channel-name="${channelName}"]`);
                    if ($row.length) {
                        // Update states-cell
                        var $states = $row.find('.states-cell .state-value');
                        if ($states.length >= 3) {
                            $states.eq(0).text(Number(stats.in_flight).toLocaleString());
                            $states.eq(1).text(Number(stats.requeued).toLocaleString());
                            $states.eq(2).text(Number(stats.deferred).toLocaleString());
                        }
                        // Update messages-cell
                        var $messages = $row.find('.messages-cell .message-value');
                        if ($messages.length >= 2) {
                            $messages.eq(0).text(Number(stats.depth).toLocaleString());
                            $messages.eq(1).text(Number(stats.messages).toLocaleString());
                        }
                        // Update consumer count section
                        var $consumerDiv = $row.find('.channel-consumer-count');
                        if ($consumerDiv.length) {
                            $consumerDiv.text('Consumers: ' + (typeof stats.consumer_count !== 'undefined' ? stats.consumer_count : '-'));
                        }
                    }
                });
            }
        }).fail(function() {
            $('.topic-stats-depth').text('-');
            $('.topic-stats-messages').text('-');
        });
    }
    window.fetchAndUpdateStats = fetchAndUpdateStats;

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
                // if (autorefreshCountdown % 2 === 0) {
                    refreshStats();
                // }
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
        if (publishVisible || tailVisible) {
            $('#main-panel').addClass('shrink-panel');
        } else {
            $('#main-panel').removeClass('shrink-panel');
        }
    }

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
            nsqd_hosts: (currentTopicDetail.nsqd_hosts || []).map(function(host) {
                if (typeof host === 'object' && host.address) {
                    return host.address;
                }
                return host;
            })
        };
        $.ajax({
            url: '/api/topic/publish?entity_id=' + currentTopicDetail.id,
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

function getTopicIDFromURL() {
    var params = new URLSearchParams(window.location.search);
    return params.get('id');
} 