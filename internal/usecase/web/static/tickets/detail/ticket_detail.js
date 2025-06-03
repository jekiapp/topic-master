$(function() {
    // Helper to get query param
    function getQueryParam(name) {
        // Support URLs like http://localhost:4181/#ticket-detail?id=xxx
        let query = '';
        if (window.parent.location.hash && window.parent.location.hash.indexOf('?') !== -1) {
            query = window.parent.location.hash.substring(window.parent.location.hash.indexOf('?'));
        } else {
            query = window.parent.location.search;
        }
        const params = new URLSearchParams(query);
        return params.get(name);
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
            // Unwrap if response is wrapped in {status, message, data}
            if (data && data.data) data = data.data;

            // Title and created time
            $('#detail-title').text(data.ticket.title || data.ticket.id);
            if (data.created_at) {
                $('#created-time').text('Created at: ' + data.created_at);
            } else {
                $('#created-time').text('');
            }
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
                    $historyTbody.append(
                        '<tr>' +
                        '<td>' + h.action + '</td>' +
                        '<td>' + (h.comment || '-') + '</td>' +
                        '<td>' +
                            h.actor + '<br>' +
                            '<span class="history-meta">' + h.created_at + '</span>' +
                        '</td>' +
                        '</tr>'
                    );
                });
            } else {
                $historyTbody.append('<tr><td colspan="3">-</td></tr>');
            }
        },
        error: function(xhr) {
            alert('Failed to load ticket detail: ' + (xhr.responseText || xhr.status));
        }
    });

    // Back link handler
    $('#back-link').on('click', function(e) {
        e.preventDefault();
        window.parent.location.hash = '#tickets';
    });
}); 