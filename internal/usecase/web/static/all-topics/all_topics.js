function formatDate(ts) {
  if (!ts) return '';
  const d = new Date(ts * 1000);
  return d.toLocaleString();
}
function renderBookmark(isBookmarked) {
  if (isBookmarked) {
    return `<img class="bookmark-icon active" src="/icons/bookmark-40px.png" alt="Bookmarked" style="width:16px;height:23px;vertical-align:middle;cursor:pointer;" />`;
  } else {
    return `<img class="bookmark-icon" src="/icons/bookmark-grey-40px.png" alt="Not Bookmarked" style="width:16px;height:23px;vertical-align:middle;cursor:pointer;" />`;
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
        return `<tr class="topic-row" data-id="${t.id}" data-bookmarked="${t.bookmarked}">
          <td>${t.name || ''}</td>
          <td>${t.group_owner || ''}</td>
          <td>${t.event_trigger || ''}</td>
          <td style="text-align:center;vertical-align:middle;">${renderBookmark(t.bookmarked)}</td>
        </tr>`;
      }).join('');
      $('#topics-table tbody').html(rows || '<tr><td colspan="7">No topics found.</td></tr>');
      // Add click handler for rows
      $('#topics-table').off('click', '.topic-row').on('click', '.topic-row', function(e) {
        // Prevent row click if bookmark icon was clicked
        if ($(e.target).hasClass('bookmark-icon')) return;
        const id = $(this).data('id');
        if (id) {
            window.parent.location.hash = `topic-detail?id=${id}`;
        }
      });
      // Add click handler for bookmark icon
      $('#topics-table').off('click', '.bookmark-icon').on('click', '.bookmark-icon', function(e) {
        e.preventDefault();
        e.stopPropagation();
        const $row = $(this).closest('.topic-row');
        const id = $row.data('id');
        let bookmarked = $row.data('bookmarked');
        if (!window.parent.isLogin || !window.parent.isLogin()) {
          alert('Please log in to bookmark topics.');
          return;
        }
        $.ajax({
          url: '/api/entity/toggle-bookmark',
          method: 'POST',
          contentType: 'application/json',
          data: JSON.stringify({ entity_id: id, bookmark: !bookmarked }),
          success: () => {
            // Toggle state locally for instant feedback
            bookmarked = !bookmarked;
            $row.data('bookmarked', bookmarked);
            $(this).replaceWith(renderBookmark(bookmarked));
          },
          error: function(xhr) {
            var msg = 'Failed to toggle bookmark';
            if (xhr.responseJSON && xhr.responseJSON.error) {
              msg += ': ' + xhr.responseJSON.error;
            }
            alert(msg);
          }
        });
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