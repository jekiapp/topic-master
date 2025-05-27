$(document).ready(function() {
    // Static group list for demo
    const groups = [
        { id: 1, name: 'admin' },
        { id: 2, name: 'user' },
        { id: 3, name: 'guest' },
        { id: 4, name: 'dev' },
        { id: 5, name: 'ops' }
    ];
    window._groupNameToId = {};
    window._groupIdToName = {};
    const $datalist = $('#signup-group-list');
    $datalist.empty();
    groups.forEach(function(g) {
        $datalist.append(`<option value="${g.name}"></option>`);
        window._groupNameToId[g.name] = g.id;
        window._groupIdToName[g.id] = g.name;
    });

    // Password confirmation validation and group validation
    $('#signup-form').on('submit', function(e) {
        e.preventDefault();
        $('#signup-form-error').hide().text('');
        const password = $('#signup-password').val();
        const confirm = $('#signup-confirm-password').val();
        const groupName = $('#signup-group-name').val();
        if (password !== confirm) {
            $('#signup-form-error').text('Passwords do not match.').show();
            return false;
        }
        if (!groupName || !window._groupNameToId[groupName]) {
            $('#signup-form-error').text('Please select a valid group.').show();
            return false;
        }
        // For now, just show a success message
        alert('Signup form is valid! (No backend yet)');
        return false;
    });
}); 