let pendingDeleteGroupId = null;
let pendingDeleteGroupName = '';
let pendingDeleteUserId = null;
let pendingDeleteUsername = '';

// Store users globally for offline access
window._userListById = {};

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
      window._userListById = {};
      if (resp && resp.data && resp.data.users) {
        resp.data.users.forEach(u => {
          $tbody.append(renderUserRow(u));
          window._userListById[u.id] = u;
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

// Utility to show user success modal with custom title/message
function showUserSuccessModal(username, password, title, message) {
  $('#user-success-popup-overlay h3').text(title);
  $('#user-success-popup-overlay .form-note').text(message).show();
  $('#created-username').text(username || '');
  $('#created-password').text(password || '');
  $('#user-success-popup-overlay').show();
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
        showUserSuccessModal(
          resp.data.username || '',
          resp.data.generated_password || '',
          'User Created',
          'The user has been created successfully.'
        );
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
  // Bind edit-user button
  $(document).on('click', '.edit-user', function() {
    const $tr = $(this).closest('tr');
    const userId = $tr.data('user-id');
    showUpdateUserPopup(userId);
  });
  // Bind delete-user button
  $(document).on('click', '.delete-user', function() {
    const $tr = $(this).closest('tr');
    const userId = $tr.data('user-id');
    const username = $tr.data('username');
    showDeleteUserPopup(userId, username);
  });
  // Cancel update user button
  $('#cancel-update-user-btn').on('click', function() {
    closeUpdateUserPopup();
  });
  // Update user form submit
  $('#update-user-form').on('submit', handleUpdateUserFormSubmit);
  // Bind delete user modal buttons here
  $('#confirm-delete-user-btn').on('click', function() {
    if (!pendingDeleteUserId) return;
    $.ajax({
      url: '/api/user/delete',
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({ user_id: pendingDeleteUserId }),
      success: function(resp) {
        closeDeleteUserPopup();
        fillUsersTable();
      },
      error: function(xhr) {
        let msg = 'Failed to delete user';
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
  $('#cancel-delete-user-btn').on('click', function() {
    closeDeleteUserPopup();
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

// --- Update User Modal Logic ---
function showUpdateUserPopup(userId) {
  // Use offline user data
  const u = window._userListById[userId];
  if (!u) {
    alert('User data not found');
    return;
  }
  $('#update-user-username').val(u.username);
  $('#update-user-name').val(u.name);
  $('#update-user-reset-password').prop('checked', false);
  // Populate group datalist
  $.ajax({
    url: '/api/group/list',
    method: 'POST',
    contentType: 'application/json',
    data: '{}',
    success: function(gresp) {
      const $datalist = $('#update-group-list');
      $datalist.empty();
      window._updateGroupNameToId = {};
      window._updateGroupIdToName = {};
      window._updateAvailableGroups = [];
      if (gresp && gresp.data && gresp.data.groups) {
        gresp.data.groups.forEach(function(g) {
          $datalist.append(`<option value="${g.name}"></option>`);
          window._updateGroupNameToId[g.name] = g.id;
          window._updateGroupIdToName[g.id] = g.name;
          window._updateAvailableGroups.push({id: g.id, name: g.name});
        });
      }
      // Fill selected groups
      const $tbody = $('#update-selected-groups-table tbody');
      $tbody.empty();
      if (u.group_details && Array.isArray(u.group_details)) {
        u.group_details.forEach(function(g) {
          $tbody.append(
            `<tr data-group-id="${g.group_id}">
              <td style="padding:4px 8px;">
                ${g.group_name}
                <span class="remove-update-group-row" style="margin-left:8px;cursor:pointer;color:#d9534f;font-weight:bold;">&times;</span>
              </td>
              <td style="padding:4px 8px;">
                <select class="update-group-type-select themed-input" style="min-width:90px;">
                  <option value="member"${g.type === 'member' ? ' selected' : ''}>member</option>
                  <option value="admin"${g.type === 'admin' ? ' selected' : ''}>admin</option>
                </select>
              </td>
            </tr>`
          );
          // Remove from datalist
          $(`#update-group-list option[value="${g.group_name}"]`).remove();
        });
      }
      $('#user-update-popup-overlay').show();
      $('#user-update-form-error').hide().text('');
    }
  });
}

// Add group to update table on selection
$(document).on('change', '#update-user-group-name', function() {
  var name = $(this).val();
  var id = window._updateGroupNameToId ? window._updateGroupNameToId[name] : '';
  if (id && name) {
    if ($(`#update-selected-groups-table tbody tr[data-group-id="${id}"]`).length === 0) {
      $('#update-selected-groups-table tbody').append(
        `<tr data-group-id="${id}">
          <td style="padding:4px 8px;">
            ${name}
            <span class="remove-update-group-row" style="margin-left:8px;cursor:pointer;color:#d9534f;font-weight:bold;">&times;</span>
          </td>
          <td style="padding:4px 8px;">
            <select class="update-group-type-select themed-input" style="min-width:90px;">
              <option value="member" selected>member</option>
              <option value="admin">admin</option>
            </select>
          </td>
        </tr>`
      );
      $(`#update-group-list option[value="${name}"]`).remove();
    }
    $(this).val('');
  }
});
// Remove group row in update modal
$(document).on('click', '.remove-update-group-row', function() {
  var $row = $(this).closest('tr');
  var id = $row.data('group-id');
  var name = window._updateGroupIdToName ? window._updateGroupIdToName[id] : '';
  if (name) {
    $('#update-group-list').append(`<option value="${name}"></option>`);
  }
  $row.remove();
});
function collectUpdateGroupMappings() {
  var groups = [];
  $('#update-selected-groups-table tbody tr').each(function() {
    var group_id = $(this).data('group-id');
    var type = $(this).find('.update-group-type-select').val();
    groups.push({ group_id: group_id, type: type });
  });
  return groups;
}
function handleUpdateUserFormSubmit(e) {
  console.log('Update user form submit triggered');
  e.preventDefault();
  $('#user-update-form-error').hide().text('');
  const username = $('#update-user-username').val();
  const name = $('#update-user-name').val();
  const reset_password = $('#update-user-reset-password').is(':checked');
  const groups = collectUpdateGroupMappings();
  const payload = {
    username: username,
    name: name,
    reset_password: reset_password,
    groups: groups
  };
  console.log('Payload to /api/user/update:', payload);
  $.ajax({
    url: '/api/user/update',
    method: 'POST',
    contentType: 'application/json',
    data: JSON.stringify(payload),
    success: function(resp) {
      console.log('Update user success:', resp);
      closeUpdateUserPopup();
      if (resp && resp.data && resp.data.generated_password) {
        showUserSuccessModal(
          resp.data.username || '',
          resp.data.generated_password || '',
          'Password Reset',
          'The password has been reset for this user.'
        );
      }
      fillUsersTable();
    },
    error: function(xhr) {
      console.log('Update user error:', xhr);
      let msg = 'Failed to update user';
      if (xhr.responseJSON && xhr.responseJSON.error) {
        msg = xhr.responseJSON.error;
      } else if (xhr.responseText) {
        try {
          const data = JSON.parse(xhr.responseText);
          if (data.error) msg = data.error;
        } catch {}
      }
      $('#user-update-form-error').text(msg).show();
    },
    complete: function(xhr, status) {
      console.log('Update user request complete:', status, xhr);
    }
  });
}
function closeUpdateUserPopup() {
  $('#user-update-popup-overlay').hide();
  $('#user-update-form-error').hide().text('');
  $('#update-user-form')[0].reset();
  $('#update-selected-groups-table tbody').empty();
  $('#update-group-list').empty();
  $('#update-user-group-name').val('');
}

function showDeleteUserPopup(userId, username) {
  pendingDeleteUserId = userId;
  pendingDeleteUsername = username;
  $('#delete-user-message').text(`Are you sure you want to delete user '${username}'?`);
  $('#delete-user-popup-overlay').show();
}

function closeDeleteUserPopup() {
  pendingDeleteUserId = null;
  pendingDeleteUsername = '';
  $('#delete-user-popup-overlay').hide();
}

$(function() {
  fillGroupsTable();
  fillUsersTable();
  bindGroupEvents();
  bindUserEvents();
});
