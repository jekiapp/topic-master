let pendingDeleteGroupId = null;
let pendingDeleteGroupName = '';

function renderGroupRow(g) {
  return `
    <tr data-group-id="${g.id}" data-group-name="${g.name}" data-group-desc="${g.description}">
      <td>${g.name}</td>
      <td>${g.description}</td>
      <td>${g.members}</td>
      <td>
        <span class="action-icon edit-group" title="Edit">
          <img src="icons/edit_icon.png" alt="Edit" style="width:15px;height:18px;vertical-align:middle;" />
        </span>
        <span style="display:inline-block; width:3px;"></span>
        <span class="action-icon delete-group" title="Delete">
          <img src="icons/delete_icon.png" alt="Delete" style="width:15px;height:18px;vertical-align:middle;" />
        </span>
      </td>
    </tr>
  `;
}

function renderUserRow(u) {
  return `<tr data-user-id="${u.id}" data-username="${u.username}">
    <td>${u.username}</td>
    <td>${u.name}</td>
    <td>${u.email}</td>
    <td>${u.groups}</td>
    <td>${u.status}</td>
    <td>
      <span class="action-icon edit-user" title="Edit">
        <img src="icons/edit_icon.png" alt="Edit" style="width:15px;height:18px;vertical-align:middle;" />
      </span>
      <span style="display:inline-block; width:3px;"></span>
      <span class="action-icon delete-user" title="Delete">
        <img src="icons/delete_icon.png" alt="Delete" style="width:15px;height:18px;vertical-align:middle;" />
      </span>
    </td>
  </tr>`;
}

function fillGroupsTable() {
  const $tbody = $('#groups-tbody');
  $tbody.empty();
  $.ajax({
    url: '/api/group/list',
    method: 'POST',
    contentType: 'application/json',
    data: '{}',
    success: function(resp) {
      if (resp && resp.data && resp.data.groups) {
        resp.data.groups.forEach(g => {
          $tbody.append(renderGroupRow(g));
        });
      }
    }
  });
}

function fillUsersTable() {
  const $tbody = $('#users-tbody');
  $tbody.empty();
  $.ajax({
    url: '/api/user/list',
    method: 'POST',
    contentType: 'application/json',
    data: '{}',
    success: function(resp) {
      if (resp && resp.data && resp.data.users) {
        resp.data.users.forEach(u => {
          $tbody.append(renderUserRow(u));
        });
      }
    }
  });
}

function showGroupFormError(msg) {
  $('#group-form-error').text(msg).show();
}

function clearGroupFormError() {
  $('#group-form-error').hide().text('');
}

function resetGroupForm() {
  $('#create-group-form')[0].reset();
  clearGroupFormError();
}

function showGroupPopup({ mode, group = {} }) {
  if (mode === 'edit') {
    $('#group-popup-overlay h3').text('Edit Group');
    $('#group-name').val(group.name).prop('disabled', true);
    $('#group-desc').val(group.description);
    $('#create-group-form').data('edit-group-id', group.id);
  } else {
    $('#group-popup-overlay h3').text('Create New Group');
    $('#group-name').val('').prop('disabled', false);
    $('#group-desc').val('');
    $('#create-group-form').removeData('edit-group-id');
  }
  clearGroupFormError();
  $('#group-popup-overlay').show();
}

function closeGroupPopup() {
  $('#group-popup-overlay').hide();
  $('#group-popup-overlay h3').text('Create New Group');
  $('#group-name').prop('disabled', false);
  $('#create-group-form').removeData('edit-group-id');
  resetGroupForm();
}

function handleGroupFormSubmit(e) {
  e.preventDefault();
  clearGroupFormError();
  const name = $('#group-name').val();
  const description = $('#group-desc').val();
  const editId = $('#create-group-form').data('edit-group-id');
  if (editId) {
    // Edit mode
    $.ajax({
      url: '/api/group/update-group-by-id',
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({ id: editId, description }),
      success: function(resp) {
        closeGroupPopup();
        fillGroupsTable();
      },
      error: function(xhr) {
        let msg = 'Failed to update group';
        if (xhr.responseJSON && xhr.responseJSON.error) {
          msg = xhr.responseJSON.error;
        } else if (xhr.responseText) {
          try {
            const data = JSON.parse(xhr.responseText);
            if (data.error) msg = data.error;
          } catch {}
        }
        showGroupFormError(msg);
      }
    });
  } else {
    // Create mode
    $.ajax({
      url: '/api/group/create',
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({ name, description }),
      success: function(resp) {
        closeGroupPopup();
        fillGroupsTable();
      },
      error: function(xhr) {
        let msg = 'Failed to create group';
        if (xhr.responseJSON && xhr.responseJSON.error) {
          msg = xhr.responseJSON.error;
        } else if (xhr.responseText) {
          try {
            const data = JSON.parse(xhr.responseText);
            if (data.error) msg = data.error;
          } catch {}
        }
        showGroupFormError(msg);
      }
    });
  }
}

function showDeleteGroupPopup(groupId, groupName) {
  pendingDeleteGroupId = groupId;
  pendingDeleteGroupName = groupName;
  $('#delete-group-message').text(`Are you sure you want to delete group '${groupName}'?`);
  $('#delete-group-popup-overlay').show();
}

function closeDeleteGroupPopup() {
  pendingDeleteGroupId = null;
  pendingDeleteGroupName = '';
  $('#delete-group-popup-overlay').hide();
}

function bindGroupEvents() {
  // Create group button
  $('#create-group-btn').on('click', function() {
    showGroupPopup({ mode: 'create' });
  });

  // Cancel button
  $('#cancel-group-btn').on('click', function() {
    closeGroupPopup();
  });

  // Edit group icon
  $(document).on('click', '.edit-group', function() {
    const $tr = $(this).closest('tr');
    showGroupPopup({
      mode: 'edit',
      group: {
        id: $tr.data('group-id'),
        name: $tr.data('group-name'),
        description: $tr.data('group-desc'),
      }
    });
  });

  // Delete group icon
  $(document).on('click', '.delete-group', function() {
    const $tr = $(this).closest('tr');
    const groupName = $tr.data('group-name');
    const groupId = $tr.data('group-id');
    showDeleteGroupPopup(groupId, groupName);
  });

  // Confirm delete
  $('#confirm-delete-group-btn').on('click', function() {
    if (!pendingDeleteGroupId) return;
    $.ajax({
      url: '/api/group/delete-group',
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({ id: pendingDeleteGroupId }),
      success: function(resp) {
        closeDeleteGroupPopup();
        fillGroupsTable();
      },
      error: function(xhr) {
        let msg = 'Failed to delete group';
        if (xhr.responseJSON && xhr.responseJSON.error) {
          msg = xhr.responseJSON.error;
        } else if (xhr.responseText) {
          try {
            const data = JSON.parse(xhr.responseText);
            if (data.error) msg = data.error;
          } catch {}
        }
        alert(msg);
      }
    });
  });

  // Cancel delete
  $('#cancel-delete-group-btn').on('click', function() {
    closeDeleteGroupPopup();
  });

  // Form submit
  $('#create-group-form').on('submit', handleGroupFormSubmit);
}

function showUserPopup() {
  $('#user-popup-overlay').show();
  $('#user-form-error').hide().text('');
  $('#create-user-form')[0].reset();
  // Populate group datalist with names only
  $.ajax({
    url: '/api/group/list',
    method: 'POST',
    contentType: 'application/json',
    data: '{}',
    success: function(resp) {
      const $datalist = $('#group-list');
      $datalist.empty();
      window._groupNameToId = {};
      window._groupIdToName = {};
      window._availableGroups = [];
      if (resp && resp.data && resp.data.groups) {
        resp.data.groups.forEach(function(g) {
          $datalist.append(`<option value="${g.name}"></option>`);
          window._groupNameToId[g.name] = g.id;
          window._groupIdToName[g.id] = g.name;
          window._availableGroups.push({id: g.id, name: g.name});
        });
      }
      // Reset selected groups table
      $('#selected-groups-table tbody').empty();
    }
  });
  // Clear group name input
  $('#user-group-name').val('');
}

// Add group to table on selection
$(document).on('change', '#user-group-name', function() {
  var name = $(this).val();
  var id = window._groupNameToId ? window._groupNameToId[name] : '';
  if (id && name) {
    // Add row if not already present
    if ($(`#selected-groups-table tbody tr[data-group-id="${id}"]`).length === 0) {
      $('#selected-groups-table tbody').append(
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
      $(`#group-list option[value="${name}"]`).remove();
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
    $('#group-list').append(`<option value="${name}"></option>`);
  }
  $row.remove();
});

function collectGroupMappings() {
  var groups = [];
  $('#selected-groups-table tbody tr').each(function() {
    var group_id = $(this).data('group-id');
    var type = $(this).find('.group-type-select').val();
    groups.push({ group_id: group_id, type: type });
  });
  return groups;
}

function handleUserFormSubmit(e) {
  e.preventDefault();
  $('#user-form-error').hide().text('');
  const username = $('#user-username').val();
  const name = $('#user-name').val();
  const email = $('#user-email').val();
  const groups = collectGroupMappings();
  $.ajax({
    url: '/api/user/create',
    method: 'POST',
    contentType: 'application/json',
    data: JSON.stringify({
      username: username,
      name: name,
      email: email,
      groups: groups
    }),
    success: function(resp) {
      closeUserPopup();
      if (resp && resp.data) {
        $('#created-username').text(resp.data.username || '');
        $('#created-password').text(resp.data.generated_password || '');
        $('#user-success-popup-overlay').show();
      }
      fillUsersTable();
    },
    error: function(xhr) {
      let msg = 'Failed to create user';
      if (xhr.responseJSON && xhr.responseJSON.error) {
        msg = xhr.responseJSON.error;
      } else if (xhr.responseText) {
        try {
          const data = JSON.parse(xhr.responseText);
          if (data.error) msg = data.error;
        } catch {}
      }
      $('#user-form-error').text(msg).show();
    }
  });
}

function bindUserEvents() {
  // Create user button
  $('#create-user-btn').on('click', function() {
    showUserPopup();
  });
  // Cancel button
  $('#cancel-user-btn').on('click', function() {
    closeUserPopup();
  });
  // Form submit
  $('#create-user-form').on('submit', handleUserFormSubmit);
  // Close user success modal
  $('#close-user-success-btn').on('click', function() {
    $('#user-success-popup-overlay').hide();
    $('#created-username').text('');
    $('#created-password').text('');
  });
  // Copy password button
  $('#copy-password-btn').on('click', function() {
    const pwd = $('#created-password').text();
    if (pwd) {
      navigator.clipboard.writeText(pwd);
    }
  });
}

function closeUserPopup() {
  $('#user-popup-overlay').hide();
  $('#user-form-error').hide().text('');
  $('#create-user-form')[0].reset();
  $('#selected-groups-table tbody').empty();
  $('#group-list').empty();
  $('#user-group-name').val('');
}

$(function() {
  fillGroupsTable();
  fillUsersTable();
  bindGroupEvents();
  bindUserEvents();
});
