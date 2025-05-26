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
          $tbody.append(`<tr><td>${u.username}</td><td>${u.name}</td><td>${u.email}</td><td>${u.groups}</td><td>${u.status}</td></tr>`);
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

$(function() {
  fillGroupsTable();
  fillUsersTable();
  bindGroupEvents();
});
