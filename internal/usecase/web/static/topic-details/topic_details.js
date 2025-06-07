// Mock data based on NsqTopicDetailResponse
const mockTopicDetail = {
    id: "1",
    name: "ExampleTopic",
    event_trigger: "user.signup\nwith extra info",
    group_owner: "alice",
    bookmarked: true,
    permission: {
        can_pause: true,
        can_publish: true,
        can_tail: true,
        can_delete: true,
        can_empty_queue: false,
        can_update_event_trigger: true
    },
    nsqd_hosts: ["1234.1234:4151", "123.13:4151"],
    topic_stats: {
        depth: 100,
        messages: 5000
    }
};

$(function() {
    // Example: get topic name from URL or a global JS variable
    // For now, use a hardcoded topic name for demo
    var topicID = getTopicNameFromURL();
    if (!topicID) {
        alert('Topic name is required');
        return;
    }

    // Fetch topic detail
    $.get('/api/topic/detail', { topic: topicID }, function(response) {
        if (!response || !response.data) {
            alert('Failed to load topic detail');
            return;
        }
        var detail = response.data;
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
    var autorefreshTimer = null;
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

    function refreshStats() {
        // Use the same logic as in the main stats fetch
        var topicID = getTopicNameFromURL();
        if (!topicID) return;
        $.get('/api/topic/detail', { topic: topicID }, function(response) {
            if (!response || !response.data) return;
            var detail = response.data;
            fetchAndUpdateStats(detail);
        });
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
});

// Helper: get topic name from URL (e.g., ?topic=MyTopic)
function getTopicNameFromURL() {
    var params = new URLSearchParams(window.location.search);
    return params.get('id');
} 