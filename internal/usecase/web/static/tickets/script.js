$(document).ready(function() {
    // Fetch and display assignments
    function loadAssignments() {
        $.ajax({
            url: '/api/tickets/list-my-assignment',
            method: 'GET',
            success: function(response) {
                var tbody = $('#assignments-tbody');
                tbody.empty();
                response.data.applications.forEach(function(app) {
                    var row = $('<tr>');
                    row.append($('<td>').text(app.title));
                    row.append($('<td>').text(app.status));
                    row.append($('<td>').text(app.reason));
                    // Format created_at as HH:mm DD/MM/YYYY
                    var createdAt = new Date(app.created_at);
                    var formattedDate = createdAt.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) + ' ' +
                        createdAt.toLocaleDateString('en-GB');
                    row.append($('<td>').text(formattedDate));
                    var actionCell = $('<td>');
                    actionCell.append('<span class="action-icon view-assignment" title="View"><i class="fa fa-eye"></i></span>');
                    row.append(actionCell);
                    tbody.append(row);
                });
            },
            error: function() {
                alert('Failed to load assignments.');
            }
        });
    }

    loadAssignments();

    // Add more event handlers as needed
});