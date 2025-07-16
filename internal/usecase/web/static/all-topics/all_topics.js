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
  // Get current URL params
  const urlParams = new URLSearchParams(window.location.search);
  const isBookmarked = urlParams.get('is_bookmarked');
  // Set heading once based on isBookmarked
  if (isBookmarked !== null) {
    $("h2").text("My Topics");
  } else {
    $("h2").text("All Topics");
  }
  // If is_bookmarked is set and user is not logged in, show message and exit
  if (isBookmarked !== null && (!window.parent.isLogin || !window.parent.isLogin())) {
    $('#topics-table tbody').html('<tr><td colspan="7" style="color: var(--error-red);">Please login to see bookmarked topics</td></tr>');
    return;
  }
  let apiUrl = '/api/topic/list-all-topics';
  if (isBookmarked !== null) {
    apiUrl += '?is_bookmarked=' + encodeURIComponent(isBookmarked);
  }

  let originalTopics = [];

  function renderTopics(topics) {
    const rows = topics.map(function(t) {
      let groupOwnerCell;
      if (!t.group_owner || t.group_owner === 'None') {
        groupOwnerCell = '<span style="display:inline-block;min-width:60px;padding:2px 12px;border-radius:999px;background:#e0e0e0;color:#888;text-align:center;">None</span>';
      } else {
        groupOwnerCell = `<span style="display:inline-block;min-width:60px;padding:2px 12px;border-radius:999px;background:#d4f7d4;color:#222;text-align:center;">${t.group_owner}</span>`;
      }
      return `<tr class="topic-row" data-id="${t.id}" data-bookmarked="${t.bookmarked}">
        <td>${t.name || ''}</td>
        <td>${groupOwnerCell}</td>
        <td>${t.event_trigger || ''}</td>
        <td style="text-align:center;vertical-align:middle;">${renderBookmark(t.bookmarked)}</td>
      </tr>`;
    }).join('');
    $('#topics-table tbody').html(rows || '<tr><td colspan="7">No topics found.</td></tr>');
  }

  $.ajax({
    url: apiUrl,
    dataType: 'json',
    xhrFields: { withCredentials: true },
    success: function(resp) {
      if (resp.status === "success" && isBookmarked !== null && resp.data.topics === null) {
        $('#topics-table tbody').html(`<tr><td colspan="7" style="text-align: center;">Bookmarked topics will be shown here.</td></tr>`);
        return;
      }
      const topics = resp.data && resp.data.topics ? resp.data.topics : [];
      originalTopics = topics;
      renderTopics(topics);
      // Add click handler for rows
      $('#topics-table').off('click', '.topic-row').on('click', '.topic-row', function(e) {
        // Prevent row click if bookmark icon was clicked
        if ($(e.target).hasClass('bookmark-icon')) return;
        const id = $(this).data('id');
        if (id) {
          const back = (isBookmarked !== null) ? 'my-topics' : 'all-topics';
          window.parent.location.hash = `topic-detail?id=${id}&back=${back}`;
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
          window.parent.showModalOverlay('Please log in to bookmark topics.');
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
            window.showModalOverlay(msg);
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

  // Local search functionality
  function doLocalSearch() {
    const query = $('#search-bar').val();
    if (!query) {
      renderTopics(originalTopics);
      return;
    }
    let regex;
    try {
      regex = new RegExp(query, 'i');
    } catch (e) {
      renderTopics([]);
      return;
    }
    const filtered = originalTopics.filter(t => regex.test(t.name || ''));
    renderTopics(filtered);
  }

  $('#find-btn').on('click', function() {
    doLocalSearch();
  });

  $('#search-bar').on('input', function() {
    if (!$(this).val()) {
      renderTopics(originalTopics);
    }
  });

  $('#search-bar').on('keydown', function(e) {
    if (e.key === 'Enter') {
      doLocalSearch();
    }
  });

  $('#refresh-btn').on('click', function() {
    var $btn = $(this);
    $btn.prop('disabled', true).text('Refreshing...');
    $.ajax({
      url: '/api/sync-topics',
      method: 'GET',
      xhrFields: { withCredentials: true },
      success: function(resp) {
        if (resp.data && resp.data.success) {
          // re-fetch topics
          location.reload();
        } else {
          window.showModalOverlay(resp && resp.error ? resp.error : 'Failed to sync topics');
        }
      },
      error: function(xhr) {
        var msg = 'Failed to sync topics';
        if (xhr.responseJSON && xhr.responseJSON.error) {
          msg = xhr.responseJSON.error;
        }
        window.showModalOverlay(msg);
      },
      complete: function() {
        $btn.prop('disabled', false).text('🔄');
      }
    });
  });
}); 