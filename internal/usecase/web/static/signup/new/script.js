$(document).ready(function() {
    // Fetch group list from backend
    window._groupNameToId = {};
    window._groupIdToName = {};
    const $datalist = $('#signup-group-list');
    $datalist.empty();
    $.ajax({
        url: '/api/group/list-simple',
        method: 'POST',
        contentType: 'application/json',
        data: '{}',
        success: function(resp) {
            if (resp && resp.data && Array.isArray(resp.data.groups)) {
                resp.data.groups.forEach(function(g) {
                    $datalist.append(`<option value="${g.name}"></option>`);
                    window._groupNameToId[g.name] = g.id;
                    window._groupIdToName[g.id] = g.name;
                });
            }
        },
        error: function(xhr) {
            $datalist.append('<option value="(error loading groups)"></option>');
        }
    });

    // Password confirmation validation and group validation, and submit to backend
    $('#signup-form').on('submit', function(e) {
        e.preventDefault();
        $('#signup-form-error').hide().text('');
        $('#signup-status-message').hide().text('');
        const username = $('#signup-username').val();
        const name = $('#signup-name').val();
        const password = $('#signup-password').val();
        const confirm = $('#signup-confirm-password').val();
        const groupName = $('#signup-group-name').val();
        const groupRole = $('#signup-group-role').val();
        const groupId = window._groupNameToId[groupName];
        if (password !== confirm) {
            $('#signup-form-error').text('Passwords do not match.').show();
            return false;
        }
        if (!groupName || !groupId) {
            $('#signup-form-error').text('Please select a valid group.').show();
            return false;
        }
        if (!username || !name || !password || !confirm || !groupType) {
            $('#signup-form-error').text('Please fill in all required fields.').show();
            return false;
        }
        // Disable form
        $('#signup-form :input').prop('disabled', true);
        $.ajax({
            url: '/api/signup',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                username: username,
                name: name,
                password: password,
                confirm_password: confirm,
                group_id: groupId,
                group_role: groupRole
            }),
            success: function(resp) {
                $('#signup-form :input').prop('disabled', false);
                $('#signup-form-error').hide().text('');
                $('#signup-status-message').text('Signup submitted! Please wait for approval.').show();
                $('#signup-form')[0].reset();
            },
            error: function(xhr) {
                $('#signup-form :input').prop('disabled', false);
                let msg = 'Signup failed.';
                if (xhr.responseJSON && xhr.responseJSON.message) {
                    msg = xhr.responseJSON.message;
                } else if (xhr.responseText) {
                    try {
                        const errObj = JSON.parse(xhr.responseText);
                        if (errObj.message) msg = errObj.message;
                    } catch {}
                }
                $('#signup-form-error').text(msg).show();
            }
        });
        return false;
    });
}); 