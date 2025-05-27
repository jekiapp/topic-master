$(function() {
    // Extract token from URL
    var params = new URLSearchParams(window.location.search);
    var token = params.get('token');
    $('#token').val(token || '');
    var $usernameLabel = $('#username-label');
    var $errorLabel = $('#reset-error');
    $errorLabel.text('');
    if (!token) {
        $usernameLabel.text('Invalid or missing token.');
        return;
    }
    // Fetch username
    $.get('/api/reset-password', { token: token })
        .done(function(data) {
            var username = data.username;
            if (!username && data.data && data.data.username) {
                username = data.data.username;
            }
            if (data.error) {
                $usernameLabel.text(data.error);
            } else if (username) {
                $usernameLabel.html('Username: <b>' + username + '</b>');
            } else {
                $usernameLabel.text('Username not found');
            }
        })
        .fail(function(xhr) {
            $usernameLabel.text('Invalid or expired token.');
        });

    // Handle form submit
    $('#reset-form').on('submit', function(event) {
        event.preventDefault();
        $errorLabel.text('');
        var new_password = $('#new-password').val();
        var confirm_password = $('#confirm-password').val();
        if (new_password !== confirm_password) {
            $errorLabel.text('Passwords do not match.');
            return;
        }
        $.ajax({
            url: '/api/reset-password',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ token: token, new_password: new_password, confirm_password: confirm_password }),
        }).done(function(data) {
            var resp = data;
            if (data.data) resp = data.data;
            if (resp.success) {
                var redirect = resp.redirect || '/login';
                window.location.href = redirect;
            } else {
                $errorLabel.text(resp.error || 'Reset failed');
            }
        }).fail(function(xhr) {
            var errorMsg = 'Reset failed';
            try {
                var data = xhr.responseJSON;
                if (data && data.error) errorMsg = data.error;
            } catch (e) {}
            $errorLabel.text(errorMsg);
        });
    });
}); 