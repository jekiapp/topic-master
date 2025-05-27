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
    window._availableGroups = [];
    const $datalist = $('#signup-group-list');
    $datalist.empty();
    groups.forEach(function(g) {
        $datalist.append(`<option value="${g.name}"></option>`);
        window._groupNameToId[g.name] = g.id;
        window._groupIdToName[g.id] = g.name;
        window._availableGroups.push({id: g.id, name: g.name});
    });

    // Add group to table on selection
    $(document).on('change', '#signup-group-name', function() {
        var name = $(this).val();
        var id = window._groupNameToId ? window._groupNameToId[name] : '';
        if (id && name) {
            // Add row if not already present
            if ($(`#signup-selected-groups-table tbody tr[data-group-id="${id}"]`).length === 0) {
                $('#signup-selected-groups-table tbody').append(
                    `<tr data-group-id="${id}">
                        <td style="padding:4px 8px;">
                            ${name}
                            <span class="remove-group-row" style="margin-left:8px;cursor:pointer;color:#d9534f;font-weight:bold;">&times;</span>
                        </td>
                        <td style="padding:4px 8px;">
                            <select class="group-type-select themed-input" style="min-width:90px;">
                                <option value="member" selected>member</option>
                                <option value="admin">admin</option>
                            </select>
                        </td>
                    </tr>`
                );
                // Remove from datalist
                $(`#signup-group-list option[value="${name}"]`).remove();
            }
            // Clear input
            $(this).val('');
        }
    });

    // Remove group row
    $(document).on('click', '.remove-group-row', function() {
        var $row = $(this).closest('tr');
        var id = $row.data('group-id');
        var name = window._groupIdToName ? window._groupIdToName[id] : '';
        if (name) {
            // Re-add to datalist
            $('#signup-group-list').append(`<option value="${name}"></option>`);
        }
        $row.remove();
    });

    // Password confirmation validation
    $('#signup-form').on('submit', function(e) {
        e.preventDefault();
        $('#signup-form-error').hide().text('');
        const password = $('#signup-password').val();
        const confirm = $('#signup-confirm-password').val();
        if (password !== confirm) {
            $('#signup-form-error').text('Passwords do not match.').show();
            return false;
        }
        // For now, just show a success message
        alert('Signup form is valid! (No backend yet)');
        return false;
    });
}); 