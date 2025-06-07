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
    // Fill data
    $('.topic-name').text(mockTopicDetail.name);
    $('.group-owner').text(mockTopicDetail.group_owner);
    var $eventTrigger = $('.event-trigger-input');
    $eventTrigger.val(mockTopicDetail.event_trigger);
    $eventTrigger.prop('readonly', !mockTopicDetail.permission.can_update_event_trigger);
    $eventTrigger.data('original', mockTopicDetail.event_trigger);

    // Bookmark icon
    var bookmarkImg = $('.bookmark-img');
    if (mockTopicDetail.bookmarked) {
        bookmarkImg.attr('src', '/icons/bookmark-true.png');
    } else {
        bookmarkImg.attr('src', '/icons/bookmark-false.png');
    }

    // Render nsqd hosts
    var hostsList = $('.nsqd-hosts-list');
    hostsList.empty();
    $.each(mockTopicDetail.nsqd_hosts, function(_, host) {
        hostsList.append($('<li>').text(host));
    });

    // Enable/disable buttons based on permission
    $('.btn-pause').prop('disabled', !mockTopicDetail.permission.can_pause);
    $('.btn-publish').prop('disabled', !mockTopicDetail.permission.can_publish);
    $('.btn-tail').prop('disabled', !mockTopicDetail.permission.can_tail);
    $('.btn-delete').prop('disabled', !mockTopicDetail.permission.can_delete);
    $('.btn-drain').prop('disabled', !mockTopicDetail.permission.can_empty_queue);

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

    // Back button (optional: history.back or custom logic)
    $('#back-link').on('click', function() {
        window.history.back();
    });

    // Render topic stats
    $('.topic-stats-depth').text(mockTopicDetail.topic_stats.depth);
    $('.topic-stats-messages').text(mockTopicDetail.topic_stats.messages);
}); 