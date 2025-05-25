function fillGroupsTable() {
  const $tbody = $('#groups-tbody');
  $tbody.empty();
  $.ajax({
    url: '/api/group-list',
    method: 'POST',
    contentType: 'application/json',
    data: '{}',
    success: function(resp) {
      if (resp && resp.data && resp.data.groups) {
        resp.data.groups.forEach(g => {
          $tbody.append(`<tr><td>${g.name}</td><td>${g.description}</td><td>${g.members}</td></tr>`);
        });
      }
    }
  });
}

function fillUsersTable() {
  const $tbody = $('#users-tbody');
  $tbody.empty();
  $.ajax({
    url: '/api/user-list',
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

$(function() {
  fillGroupsTable();
  fillUsersTable();

  $('#create-group-btn').on('click', function() {
    resetGroupForm();
    $('#group-popup-overlay').show();
  });

  $('#cancel-group-btn').on('click', function() {
    $('#group-popup-overlay').hide();
    resetGroupForm();
  });

  $('#create-group-form').on('submit', function(e) {
    e.preventDefault();
    clearGroupFormError();
    const name = $('#group-name').val();
    const description = $('#group-desc').val();
    $.ajax({
      url: '/api/create-group',
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({ name, description }),
      success: function(resp) {
        $('#group-popup-overlay').hide();
        resetGroupForm();
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
  });
});
