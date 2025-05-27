function formatDate(ts) {
  if (!ts) return '';
  const d = new Date(ts * 1000);
  return d.toLocaleString();
}
function renderBookmark(isBookmarked) {
  return `<svg class="bookmark-icon${isBookmarked ? ' active' : ''}" viewBox="0 0 24 24"><path d="M6 4a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v16l-7-5-7 5V4z"/></svg>`;
}
$(function() {
  $.ajax({
    url: '/api/topic/list-topics',
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
        return `<tr>
          <td>${t.name || ''}</td>
          <td>${t.group || ''}</td>
          <td>${t.description || ''}</td>
          <td>${formatDate(t.created_at)}</td>
          <td>${t.rps_1h != null ? t.rps_1h : '-'}</td>
          <td><span class="status">${t.status || ''}</span></td>
          <td>${renderBookmark(t.bookmarked)}</td>
        </tr>`;
      }).join('');
      $('#topics-table tbody').html(rows || '<tr><td colspan="7">No topics found.</td></tr>');
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