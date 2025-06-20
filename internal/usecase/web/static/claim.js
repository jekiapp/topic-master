// Global function to show claim modal for an entity and group
window.showClaimModal = function(entityId, entityName, onSubmit) {
    // Get user info from localStorage
    var user = localStorage.getItem('user');
    user = user ? JSON.parse(user) : null;
    var groups = (user && user.groups) ? user.groups : [];

    // Build dropdown options
    var groupOptions = groups.map(function(group) {
        return '<option value="' + group + '">' + group + '</option>';
    }).join('');

    var modalHtml = `
        <div style="min-width:300px;">
            <h3 style="margin-top:0;">Claim ${entityName} for group</h3>
            <div style="margin-bottom:1em;">
                <label for="claim-group-select">Select Group:</label>
                <select id="claim-group-select" style="width:100%;margin-top:0.5em;">
                    ${groupOptions}
                </select>
            </div>
            <div style="text-align:right;">
                <button id="claim-cancel-btn" style="margin-right:0.5em;">Cancel</button>
                <button id="claim-submit-btn">Submit</button>
            </div>
        </div>
    `;

    window.showModalOverlay(modalHtml);

    $('#claim-cancel-btn').on('click', function() {
        window.hideModalOverlay();
    });

    $('#claim-submit-btn').on('click', function() {
        var selectedGroup = $('#claim-group-select').val();
        if (typeof onSubmit === 'function') {
            onSubmit({ entityId, entityName, group: selectedGroup });
        }
        window.hideModalOverlay();
    });
};
