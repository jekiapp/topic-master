$(function() {
  // Check login status before making AJAX call
  var isLogin = (window.parent && window.parent.isLogin) ? window.parent.isLogin() : (window.isLogin && window.isLogin());
  if (!isLogin) {
    $('#topics-table tbody').html('<tr><td colspan="7" style="color: var(--error-red);">Please login to see your bookmarked topics here</td></tr>');
    return;
  }
  $.ajax({
    url: '/api/topic/list-my-topics',
    dataType: 'json',
    xhrFields: { withCredentials: true },
    success: function(resp) {
      if (resp.error) {
        let msg = resp.error === "record not found" ? "You have no eligible topics" : resp.error;
        $('#topics-table tbody').html(`<tr><td colspan="7" style="color: var(--error-red);">${msg}</td></tr>`);
        return;
      }
      const topics = resp.data && resp.data.topics ? resp.data.topics : [];
      const rows = topics.map(function(t) {
        return `<tr class="topic-row" data-id="${t.id}">
          <td>${t.name || ''}</td>
          <td>${t.group || ''}</td>
          <td>${t.type || ''}</td>
        </tr>`;
      }).join('');
      $('#topics-table tbody').html(rows || '<tr><td colspan="7">No topics found.</td></tr>');
      // Add click handler for rows
      $('#topics-table').off('click', '.topic-row').on('click', '.topic-row', function() {
        const id = $(this).data('id');
        if (id) {
            window.parent.location.hash = `topic-detail?id=${id}&back=my-topics`;
        }
      });
    },
    error: function(xhr) {
      if (xhr.status === 401 && xhr.responseJSON && xhr.responseJSON.error) {
        $('#topics-table tbody').html(`<tr><td colspan="7" style="color: var(--error-red);">${xhr.responseJSON.error}</td></tr>`);
      } else {
        $('#topics-table tbody').html('<tr><td colspan="7" style="color: var(--error-red);">Failed to load topics.</td></tr>');
      }
    }
  });
}); 