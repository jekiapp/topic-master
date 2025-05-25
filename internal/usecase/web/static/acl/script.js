// Placeholder for ACL interactivity
console.log('ACL script loaded. Ready for future features.');

// Mock data
const groups = [
  { name: 'Admins', description: 'System administrators', members: 'alice, bob' },
  { name: 'Editors', description: 'Content editors', members: 'carol, dave' },
  { name: 'Viewers', description: 'Read-only users', members: 'eve, frank' },
];

const users = [
  { username: 'alice', name: 'Alice Smith', email: 'alice@example.com', groups: 'Admins', type: 'Admin', status: 'Active' },
  { username: 'bob', name: 'Bob Jones', email: 'bob@example.com', groups: 'Admins', type: 'Admin', status: 'Active' },
  { username: 'carol', name: 'Carol White', email: 'carol@example.com', groups: 'Editors', type: 'User', status: 'Active' },
  { username: 'dave', name: 'Dave Black', email: 'dave@example.com', groups: 'Editors', type: 'User', status: 'Inactive' },
  { username: 'eve', name: 'Eve Green', email: 'eve@example.com', groups: 'Viewers', type: 'User', status: 'Active' },
  { username: 'frank', name: 'Frank Blue', email: 'frank@example.com', groups: 'Viewers', type: 'User', status: 'Inactive' },
];

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
  users.forEach(u => {
    $tbody.append(`<tr><td>${u.username}</td><td>${u.name}</td><td>${u.email}</td><td>${u.groups}</td><td>${u.type}</td><td>${u.status}</td></tr>`);
  });
}

$(function() {
  fillGroupsTable();
  fillUsersTable();
});
