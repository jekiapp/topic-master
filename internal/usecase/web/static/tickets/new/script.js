$(function() {
    // Back link handler
    $('#back-link').on('click', function(e) {
        e.preventDefault();
        window.parent.location.hash = '#tickets';
    });

    function escapeHtml(text) {
        return $('<div>').text(text).html();
    }

    function renderForm(data) {
        const $section = $('#form-section');
        $section.empty();
        if (!data) {
            $section.append('<div class="error">Failed to load form data.</div>');
            return;
        }
        // Title
        $('#form-title').text(data.title || 'New Application');
        // Applicant
        $section.append('<div><strong>Applicant:</strong> ' +
            escapeHtml(data.applicant.name) + ' (' + escapeHtml(data.applicant.username) + ')</div>');
        // Reviewers
        if (data.reviewers && data.reviewers.length > 0) {
            const reviewers = data.reviewers.map(r => escapeHtml(r.name) + ' (' + escapeHtml(r.username) + ')').join(', ');
            $section.append('<div><strong>Reviewers:</strong> ' + reviewers + '</div>');
        }
        // Form fields
        const $form = $('<form id="new-app-form"></form>');
        if (data.fields && data.fields.length > 0) {
            data.fields.forEach(function(field, idx) {
                const id = 'field-' + idx;
                let inputHtml = '';
                const required = field.required ? 'required' : '';
                const editable = field.editable ? '' : 'readonly';
                if (field.type === 'text') {
                    inputHtml = `<input type="text" id="${id}" name="${escapeHtml(field.label)}" value="${escapeHtml(field.default_value)}" ${required} ${editable} class="form-input">`;
                } else if (field.type === 'textarea') {
                    inputHtml = `<textarea id="${id}" name="${escapeHtml(field.label)}" ${required} ${editable} class="form-input">${escapeHtml(field.default_value)}</textarea>`;
                }
                $form.append(
                    `<div class="form-group">
                        <label for="${id}">${escapeHtml(field.label)}${field.required ? ' <span style="color:red">*</span>' : ''}</label><br>
                        ${inputHtml}
                    </div>`
                );
            });
        }
        // Permissions
        if (data.permissions && data.permissions.length > 0) {
            $form.append('<div class="form-group"><strong>Permissions:</strong></div>');
            data.permissions.forEach(function(p, idx) {
                const id = 'perm-' + idx;
                $form.append(
                    `<div class="form-permission">
                        <input type="checkbox" id="${id}" name="permissions" value="${escapeHtml(p.name)}">
                        <label for="${id}">${escapeHtml(p.name)} <span style="color:#888">(${escapeHtml(p.description)})</span></label>
                    </div>`
                );
            });
        }
        // Submit button
        $form.append('<button type="submit" id="submit-btn">Submit</button>');
        $section.append($form);
    }

    // Fetch form data
    $.ajax({
        url: '/api/tickets/new-application-form',
        method: 'GET',
        success: function(resp) {
            if (resp && resp.data) {
                renderForm(resp.data);
            } else {
                renderForm(null);
            }
        },
        error: function() {
            renderForm(null);
        }
    });

    // (Submission logic can be added here if needed)
}); 