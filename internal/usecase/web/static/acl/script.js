
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

$(function() {
  fillGroupsTable();
  fillUsersTable();
});
