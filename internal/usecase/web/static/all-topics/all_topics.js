function formatDate(ts) {
  if (!ts) return '';
  const d = new Date(ts * 1000);
  return d.toLocaleString();
}
function renderBookmark(isBookmarked) {
  if (isBookmarked) {
    return `<img class="bookmark-icon active" src="/icons/bookmark-40px.png" alt="Bookmarked" style="width:16px;height:23px;vertical-align:middle;" />`;
  } else {
    return `<img class="bookmark-icon" src="/icons/bookmark-grey-40px.png" alt="Not Bookmarked" style="width:16px;height:23px;vertical-align:middle;" />`;
  }
}
$(function() {
  $.ajax({
    url: '/api/topic/list-all-topics',
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
          <td>${t.group_owner || ''}</td>
          <td>${t.event_trigger || ''}</td>
          <td style="text-align:center;vertical-align:middle;">${renderBookmark(t.bookmarked)}</td>
        </tr>`;
      }).join('');
      $('#topics-table tbody').html(rows || '<tr><td colspan="7">No topics found.</td></tr>');
      // Add click handler for rows
      $('#topics-table').off('click', '.topic-row').on('click', '.topic-row', function() {
        const id = $(this).data('id');
        if (id) {
            window.parent.location.hash = `topic-detail?id=${id}`;
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