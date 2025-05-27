// Global AJAX 401 handler
$(document).ajaxError(function(event, jqxhr, settings, thrownError) {
    if (jqxhr.status === 401) {
        // You can customize this action (redirect, alert, etc)
        alert('Session expired or unauthorized. Please log in again.');
        // Optionally, redirect to login page:
        window.top.location.href = '/login';
        
    }
}); 