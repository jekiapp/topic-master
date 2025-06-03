$(function() {
    // Helper to get query param
    function getQueryParam(name) {
        const url = new URL(window.location.href);
        return url.searchParams.get(name);
    }

    const ticketId = getQueryParam('id');
    if (!ticketId) {
        alert('No ticket id specified');
        return;
    }

    $.ajax({
        url: '/api/tickets/detail?id=' + encodeURIComponent(ticketId),
        method: 'GET',
        success: function(data) {
            // Title and created time
            $('#detail-title').text(data.ticket.title || data.ticket.id);
            $('#created-time').text(data.created_at || '');
            // Applicant
            $('#detail-applicant').text(data.applicant.name + ' (' + data.applicant.username + ')');
            // Permissions
            const $permsList = $('#detail-permissions-list');
            $permsList.empty();
            if (data.ticket.permissions && data.ticket.permissions.length > 0) {
                data.ticket.permissions.forEach(function(p) {
                    const desc = p.description ? ' <span style="color:#888">(' + p.description + ')</span>' : '';
                    $permsList.append('<li>' + p.name + desc + '</li>');
                });
            } else {
                $permsList.append('<li>-</li>');
            }
            // Reason
            $('#detail-reason').text(data.ticket.reason || '-');
            // Status
            $('#detail-status').text(data.ticket.status || '-');

            // Assignees
            const $assigneeTbody = $('#assignee-table tbody');
            $assigneeTbody.empty();
            if (data.assignees && data.assignees.length > 0) {
                data.assignees.forEach(function(a) {
                    $assigneeTbody.append('<tr><td>' + a.name + ' (' + a.username + ')</td><td>' + a.status + '</td></tr>');
                });
            } else {
                $assigneeTbody.append('<tr><td colspan="2">-</td></tr>');
            }

            // Histories
            const $historyTbody = $('#history-table tbody');
            $historyTbody.empty();
            if (data.histories && data.histories.length > 0) {
                data.histories.forEach(function(h) {
                    $historyTbody.append('<tr><td>' + h.action + '</td><td>' + (h.comment || '-') + '</td><td>' + (h.created_at ? new Date(h.created_at).toLocaleString() : '-') + '</td></tr>');
                });
            } else {
                $historyTbody.append('<tr><td colspan="3">-</td></tr>');
            }
        },
        error: function(xhr) {
            alert('Failed to load ticket detail: ' + (xhr.responseText || xhr.status));
        }
    });
}); 