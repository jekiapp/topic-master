$(document).ready(function() {
    // Pagination state for assignments
    var assignmentsPage = 1;
    var assignmentsLimit = 10;
    var assignmentsHasNext = false;

    // Pagination state for my applications
    var applicationsPage = 1;
    var applicationsLimit = 10;
    var applicationsHasNext = false;

    // Fetch and display assignments
    function loadAssignments() {
        var isLogin = (window.parent && window.parent.isLogin) ? window.parent.isLogin() : (window.isLogin && window.isLogin());
        if (!isLogin) {
            var tbody = $('#assignments-tbody');
            tbody.empty();
            var row = $('<tr>').append(
                $('<td colspan="5" style="text-align:center;color: var(--error-red);">').text('Please login to see your assignments here')
            );
            tbody.append(row);
            // Update pagination controls
            $('#assignments-page').text('');
            $('#assignments-prev').prop('disabled', true);
            $('#assignments-next').prop('disabled', true);
            return;
        }
        $.ajax({
            url: '/api/tickets/list-my-assignment',
            method: 'GET',
            data: { page: assignmentsPage, limit: assignmentsLimit },
            success: function(response) {
                var tbody = $('#assignments-tbody');
                tbody.empty();
                var apps = response.data.applications;
                assignmentsHasNext = !!response.data.has_next;
                if (!Array.isArray(apps) || apps.length === 0) {
                    var emptyRow = $('<tr>').append(
                        $('<td colspan="5" style="text-align:center;">').text('No assignments found')
                    );
                    tbody.append(emptyRow);
                } else {
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
                }
                // Update pagination controls
                $('#assignments-page').text('Page ' + assignmentsPage);
                $('#assignments-prev').prop('disabled', assignmentsPage === 1);
                $('#assignments-next').prop('disabled', !assignmentsHasNext);
            },
            error: function() {
                window.parent.showModalOverlay('Failed to load assignments.');
            }
        });
    }

    // Fetch and display my applications
    function loadMyApplications() {
        var isLogin = (window.parent && window.parent.isLogin) ? window.parent.isLogin() : (window.isLogin && window.isLogin());
        if (!isLogin) {
            var tbody = $('#applications-tbody');
            tbody.empty();
            var row = $('<tr>').append(
                $('<td colspan="5" style="text-align:center;color: var(--error-red);">').text('Please login to see your applications here')
            );
            tbody.append(row);
            // Update pagination controls
            $('#applications-page').text('');
            $('#applications-prev').prop('disabled', true);
            $('#applications-next').prop('disabled', true);
            return;
        }
        $.ajax({
            url: '/api/tickets/list-my-applications',
            method: 'GET',
            data: { page: applicationsPage, limit: applicationsLimit },
            success: function(response) {
                var tbody = $('#applications-tbody');
                tbody.empty();
                var apps = response.data.applications;
                applicationsHasNext = !!response.data.has_next;
                if (!Array.isArray(apps) || apps.length === 0) {
                    var emptyRow = $('<tr>').append(
                        $('<td colspan="5" style="text-align:center;">').text('No applications found')
                    );
                    tbody.append(emptyRow);
                } else {
                    apps.forEach(function(app) {
                        var row = $('<tr>').attr('data-app-id', app.id);
                        row.append($('<td>').text(app.title));
                        row.append($('<td>').text(app.status));
                        row.append($('<td>').text(app.reason));
                        // Limit assignee names to 2, then add 'and x more' if needed
                        var assignees = (app.assignee_names || '').split(',').map(function(s) { return s.trim(); }).filter(Boolean);
                        var assigneeDisplay = '';
                        if (assignees.length > 2) {
                            assigneeDisplay = assignees.slice(0,2).join(', ') + ' and ' + (assignees.length - 2) + ' more';
                        } else {
                            assigneeDisplay = assignees.join(', ');
                        }
                        row.append($('<td>').text(assigneeDisplay));
                        // Format created_at as HH:mm DD/MM/YYYY
                        var createdAt = new Date(app.created_at);
                        var formattedDate = createdAt.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) + ' ' +
                            createdAt.toLocaleDateString('en-GB');
                        row.append($('<td>').addClass('created-at-cell').text(formattedDate));
                        tbody.append(row);
                    });
                }
                // Update pagination controls
                $('#applications-page').text('Page ' + applicationsPage);
                $('#applications-prev').prop('disabled', applicationsPage === 1);
                $('#applications-next').prop('disabled', !applicationsHasNext);
            },
            error: function() {
                window.parent.showModalOverlay('Failed to load applications.');
            }
        });
    }

    // Pagination button handlers
    $('#assignments-prev').on('click', function() {
        if (assignmentsPage > 1) {
            assignmentsPage--;
            loadAssignments();
        }
    });
    $('#assignments-next').on('click', function() {
        if (assignmentsHasNext) {
            assignmentsPage++;
            loadAssignments();
        }
    });

    // Pagination button handlers for my applications
    $('#applications-prev').on('click', function() {
        if (applicationsPage > 1) {
            applicationsPage--;
            loadMyApplications();
        }
    });
    $('#applications-next').on('click', function() {
        if (applicationsHasNext) {
            applicationsPage++;
            loadMyApplications();
        }
    });

    loadAssignments();
    loadMyApplications();

    // Make each row clickable for assignments
    $('#assignments-tbody').on('click', 'tr', function() {
        if ($(this).is(':hidden')) return;
        var appId = $(this).data('app-id');
        if (appId) {
            window.parent.location.hash = '#ticket-detail?id=' + encodeURIComponent(appId);
        }
    });

    // Make each row clickable for my applications
    $('#applications-tbody').on('click', 'tr', function() {
        if ($(this).is(':hidden')) return;
        var appId = $(this).data('app-id');
        if (appId) {
            window.parent.location.hash = '#ticket-detail?id=' + encodeURIComponent(appId);
        }
    });

    // Add more event handlers as needed
});