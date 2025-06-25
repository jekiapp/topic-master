$(document).ready(function() {
    // Fetch and display assignments
    function loadAssignments() {
        $.ajax({
            url: '/api/tickets/list-my-assignment',
            method: 'GET',
            success: function(response) {
                var tbody = $('#assignments-tbody');
                tbody.empty();
                var apps = response.data.applications;
                if (!Array.isArray(apps) || apps.length === 0) {
                    var emptyRow = $('<tr>').append(
                        $('<td colspan="4" style="text-align:center;">').text('No assignments found')
                    );
                    tbody.append(emptyRow);
                    return;
                }
                apps.forEach(function(app) {
                    var row = $('<tr>').attr('data-app-id', app.id);
                    row.append($('<td>').text(app.title));
                    row.append($('<td>').text(app.status));
                    row.append($('<td>').text(app.reason));
                    row.append($('<td>').text(app.applicant_name));
                    // Format created_at as HH:mm DD/MM/YYYY
                    var createdAt = new Date(app.created_at);
                    var formattedDate = createdAt.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) + ' ' +
                        createdAt.toLocaleDateString('en-GB');
                    row.append($('<td>').addClass('created-at-cell').text(formattedDate));
                    tbody.append(row);
                });
            },
            error: function() {
                alert('Failed to load assignments.');
            }
        });
    }

    loadAssignments();

    // Make each row clickable
    $('#assignments-tbody').on('click', 'tr', function() {
        if ($(this).is(':hidden')) return;
        var appId = $(this).data('app-id');
        if (appId) {
            window.parent.location.hash = '#ticket-detail?id=' + encodeURIComponent(appId);
        }
    });

    // Add more event handlers as needed
});