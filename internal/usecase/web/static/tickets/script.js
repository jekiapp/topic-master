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
                    row.append($('<td>').text(app.id));
                    row.append($('<td>').text(app.title));
                    row.append($('<td>').text(app.status));
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