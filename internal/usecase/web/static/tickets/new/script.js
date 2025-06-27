$(function() {
    // Back link handler
    $('#back-link').on('click', function(e) {
        e.preventDefault();
        window.parent.location.hash = '#tickets';
    });

    function escapeHtml(text) {
        return $('<div>').text(text).html();
    }

    // Helper to make a safe HTML id from a label
    function makeIdFromLabel(label) {
        return label.replace(/[^a-zA-Z0-9_-]/g, '_');
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

        // Start Application Info section
        let appInfoHtml = '<div class="section"><div class="section-title">Application Info</div>';
        // Applicant
        appInfoHtml += '<div class="form-group"><strong>Applicant:</strong> ' +
            escapeHtml(data.applicant.name) + ' (' + escapeHtml(data.applicant.username) + ')</div>';
        // Form fields
        if (data.fields && data.fields.length > 0) {
            data.fields.forEach(function(field, idx) {
                const id = makeIdFromLabel(field.label);
                let inputHtml = '';
                const required = field.required ? 'required' : '';
                const editable = field.editable ? '' : 'readonly';
                if (field.type === 'text') {
                    inputHtml = `<input type="text" id="${id}" name="${escapeHtml(field.label)}" value="${escapeHtml(field.default_value)}" ${required} ${editable} class="form-input">`;
                } else if (field.type === 'textarea') {
                    inputHtml = `<textarea id="${id}" name="${escapeHtml(field.label)}" ${required} ${editable} class="form-input">${escapeHtml(field.default_value)}</textarea>`;
                } else if (field.type === 'label') {
                    inputHtml = `<span class=\"form-label\" id=\"${id}\" style=\"color:#555;margin-left:8px;\">${escapeHtml(field.default_value)}</span>`;
                } else if (field.type === 'label-multiline') {
                    inputHtml = `<div class=\"form-label-multiline\" id=\"${id}\" style=\"white-space:pre-line;padding:8px 0 4px 0;color:#555;\">${escapeHtml(field.default_value)}</div>`;
                }
                if (field.type === 'label') {
                    appInfoHtml +=
                        `<div class=\"form-group\">\n                            <label for=\"${id}\"><strong>${escapeHtml(field.label)}</strong> : ${inputHtml}</label>\n                        </div>`;
                } else {
                    appInfoHtml +=
                        `<div class=\"form-group\">\n                            <label for=\"${id}\"><strong>${escapeHtml(field.label)}</strong> : </label><br>\n                            ${inputHtml}\n                        </div>`;
                }
            });
        }
        appInfoHtml += '</div>';
        $section.append(appInfoHtml);

        // Reviewers section
        if (data.reviewers && data.reviewers.length > 0) {
            const reviewers = data.reviewers.map(r => escapeHtml(r.name) + ' (' + escapeHtml(r.username) + ')').join(', ');
            $section.append(
                `<div class="section">
                    <div class="section-title">Reviewers</div>
                    <div>${reviewers}</div>
                </div>`
            );
        }

        // Permissions section
        if (data.permissions && data.permissions.length > 0) {
            let permHtml = `<div class="section">
                <div class="section-title">Permissions</div>`;
            data.permissions.forEach(function(p, idx) {
                const id = 'perm-' + idx;
                permHtml +=
                    `<div class="form-permission">
                        <input type="checkbox" id="${id}" name="permissions" value="${escapeHtml(p.name)}">
                        <label for="${id}">${escapeHtml(p.name)} <span style="color:#888">(${escapeHtml(p.description)})</span></label>
                    </div>`;
            });
            permHtml += '</div>';
            $section.append(permHtml);
        }
        // Submit button
        $section.append('<button type="button" id="submit-btn">Submit</button>');
    }

    // Helper to extract query params from parent hash
    function getQueryParamsFromParentHash() {
        let hash = '';
        if (window.parent && window.parent.location && window.parent.location.hash) {
            hash = window.parent.location.hash;
        } else {
            hash = window.location.hash;
        }
        const queryIndex = hash.indexOf('?');
        let params = {};
        if (queryIndex !== -1) {
            const queryString = hash.substring(queryIndex + 1);
            const urlParams = new URLSearchParams(queryString);
            for (const [key, value] of urlParams.entries()) {
                params[key] = value;
            }
        }
        return params;
    }

    // Fetch form data
    const params = getQueryParamsFromParentHash();
    let apiUrl = '/api/tickets/new-application-form';
    const queryArr = [];
    if (params.type) queryArr.push('type=' + encodeURIComponent(params.type));
    if (params.entity_id) queryArr.push('entity_id=' + encodeURIComponent(params.entity_id));
    if (queryArr.length > 0) apiUrl += '?' + queryArr.join('&');
    $.ajax({
        url: apiUrl,
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
    // --- Submission logic start ---
    $(document).on('click', '#submit-btn', function() {
        // Collect entity_id and application_type from query params
        const params = getQueryParamsFromParentHash();
        const entityId = params.entity_id || '';
        const applicationType = params.type || 'topic';
        // Collect reason by label-based id
        const reason = $('#Reason').val() || '';
        // Collect permissions from checked checkboxes by name
        const permissions = $("input[type='checkbox'][name='permissions']:checked").map(function() {
            return this.value;
        }).get();
        // Basic validation
        if (!entityId || !reason || permissions.length === 0) {
            alert('Please provide a reason and select at least one permission.');
            return;
        }
        // Prepare payload
        const payload = {
            entity_id: entityId,
            application_type: applicationType,
            reason: reason,
            permission: permissions
        };
        // POST to backend
        $.ajax({
            url: '/api/tickets/submit-application',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(payload),
            success: function(resp) {
                if (resp && resp.app_url) {
                    window.parent.location.hash = resp.app_url;
                } else {
                    alert('Submission succeeded but no redirect URL returned.');
                }
            },
            error: function(xhr) {
                let msg = 'Submission failed.';
                if (xhr && xhr.responseJSON && xhr.responseJSON.error) {
                    msg += '\n' + xhr.responseJSON.error;
                }
                alert(msg);
            }
        });
    });
    // --- Submission logic end ---
}); 