function getQueryParam(name) {
    const url = new URL(window.location.href);
    return url.searchParams.get(name);
}

$(function() {
    const id = getQueryParam('id');
    // Session check
    $.ajax({
        url: '/api/user/get-username',
        method: 'GET',
        dataType: 'json',
    }).done(function(resp) {
        // If authorized, redirect
        if (id) {
            window.location.href = '/#ticket-detail?id=' + encodeURIComponent(id);
        }
    }).fail(function(jqxhr) {
        if (jqxhr.status === 401) {
            // Unauthorized: do nothing
            return;
        }
    });
    if (!id) {
        $('#application-section').text('No application ID provided.');
        return;
    }
    $.get(`/api/signup/app`, { id: id })
        .done(function(response) {
            var data = response.data || response;
            // Application Detail
            $('#detail-title').text(data.application.title || '');
            $('#detail-username').text(data.user.username || '');
            $('#detail-name').text(data.user.name || '');
            let group = (data.user.groups && data.user.groups.length > 0) ? data.user.groups[0] : {};
            $('#detail-groupname').text(group.group_name || '');
            $('#detail-grouprole').text(group.role || '');
            $('#detail-reason').text(data.application.reason || '');
            $('#detail-status').text(data.application.status || '');
            $('#detail-createdat').text(formatDateTime(data.application.created_at));

            // Assignees
            let assignees = data.assignee || [];
            let $assigneeTbody = $('#assignee-table tbody');
            $assigneeTbody.empty();
            if (assignees.length === 0) {
                $assigneeTbody.append('<tr><td colspan="2">No assignees.</td></tr>');
            } else {
                assignees.forEach(function(a) {
                    $assigneeTbody.append(`<tr><td>${a.name || a.username || ''}</td><td>${a.status || ''}</td></tr>`);
                });
            }

            // Histories
            let histories = data.histories || [];
            let $historyTbody = $('#history-table tbody');
            $historyTbody.empty();
            if (histories.length === 0) {
                $historyTbody.append('<tr><td colspan="2">No history.</td></tr>');
            } else {
                histories.forEach(function(h) {
                    $historyTbody.append(`<tr><td>${h.action || ''}</td><td>${h.comment || ''}</td><td>${formatDateTime(h.created_at) || ''}</td></tr>`);
                });
            }
        })
        .fail(function() {
            $('#application-section').text('Failed to load application detail.');
        });
});

function formatDateTime(dt) {
    if (!dt) return '';
    var d = new Date(dt);
    if (isNaN(d.getTime())) return dt;
    return String(d.getHours()).padStart(2,'0') + ':' +
           String(d.getMinutes()).padStart(2,'0') + ' ' +
           String(d.getDate()).padStart(2,'0') + '/' +
           String(d.getMonth()+1).padStart(2,'0') + '/' +
           d.getFullYear();
} 