$(function() {
  // Session check and user info update
  function checkSessionAndUpdateUser() {
    return $.ajax({
      url: '/api/user/get-username',
      method: 'GET',
      dataType: 'json',
    }).done(function(resp) {
      var userData = resp && resp.data;
      if (userData && userData.name) {
        localStorage.setItem('user', JSON.stringify(userData));
        $('#user-dropdown').show();
        $('#login-signup-link').hide();
        $('#user-dropdown .user-name').text(userData.name + ' ▼');
      } else {
        localStorage.removeItem('user');
        $('#user-dropdown').hide();
        $('#login-signup-link').show();
      }
      // Show Access Control only if root
      if (userData && userData.root === true) {
        $(".menu li a.hidden").removeClass('hidden');
      }
    }).fail(function(jqxhr) {
      localStorage.removeItem('user');
      if (jqxhr.status === 401) {
        $('#user-dropdown').hide();
        $('#login-signup-link').show();
        // No redirect
      }
    });
  }

  // Call on page load
  checkSessionAndUpdateUser();

  const iframeContainer = $('.iframe-container');
  const myTopicsMenu = $('.menu li a').filter(function() {
    return $(this).text().trim() === 'My Topics';
  });
  const accessControlMenu = $('.menu li a').filter(function() {
    return $(this).text().trim() === 'Access Control';
  });

  const mainIframe = $('#main-iframe');

  function showMyTopics() {
    mainIframe.attr('src', 'all-topics/index.html?is_bookmarked=true');
  }

  function showAccessControl() {
    mainIframe.attr('src', 'acl/index.html');
  }

  const allTopicsMenu = $('.menu li a').filter(function() {
    return $(this).text().trim() === 'All Topics';
  });

  function showAllTopics() {
    mainIframe.attr('src', 'all-topics/index.html');
  }

  function showTopicDetail(id) {
    mainIframe.attr('src', `topic-details/index.html?id=${id}`);
  }

  function setActiveMenuByHash(hash) {
    if (hash === '#access') {
      $('.menu li a').removeClass('active');
      accessControlMenu.addClass('active');
    } else if (hash === '#tickets' || hash.startsWith('#ticket-detail')) {
    $('.menu li a').removeClass('active');
      ticketsMenu.addClass('active');
    } else if (hash === '#all-topics'|| hash === '') {
      $('.menu li a').removeClass('active');
      allTopicsMenu.addClass('active');
    } else if (hash === '#my-topics') {
      $('.menu li a').removeClass('active');
      myTopicsMenu.addClass('active');
    } 
  }

  const ticketsMenu = $('.menu li a').filter(function() {
    return $(this).text().trim() === 'Tickets';
  });

  function showTickets() {
    mainIframe.attr('src', 'tickets/index.html');
  }

  function showTicketDetail() {
    mainIframe.attr('src', 'tickets/detail/index.html');
  }

  function showTicketsNew() {
    mainIframe.attr('src', 'tickets/new/index.html');
  }

  // Add event listener for Tickets menu
  $(window).on('hashchange', handleHashChange);

  ticketsMenu.on('click', function(e) {
    e.preventDefault();
    window.location.hash = '#tickets';
    $('.menu li a').removeClass('active');
    $(this).addClass('active');
    showTickets();
  });

  // Update handleHashChange function
  function handleHashChange() {
    const hash = window.location.hash;
    setActiveMenuByHash(hash);
    if (hash === '#access') {
      showAccessControl();
    } else if (hash === '#tickets') {
      showTickets();
    } else if (hash.startsWith('#tickets-new')) {
      showTicketsNew();
    } else if (hash.startsWith('#ticket-detail')) {
      showTicketDetail();
    } else if (hash === '#my-topics') {
      showMyTopics();
    } else if (hash.startsWith('#topic-detail')) {
      showTopicDetail(hash.split('=')[1]);
    } else {
      showAllTopics();
    }
  }

  // On page load, handle hash
  handleHashChange();

  // Listen for hash changes
  $(window).on('hashchange', handleHashChange);

  myTopicsMenu.on('click', function(e) {
    e.preventDefault();
    window.location.hash = '#my-topics';
    $('.menu li a').removeClass('active');
    $(this).addClass('active');
    showMyTopics();
  });

  accessControlMenu.on('click', function(e) {
    e.preventDefault();
    window.location.hash = '#access';
    $('.menu li a').removeClass('active');
    $(this).addClass('active');
    showAccessControl();
  });

  allTopicsMenu.on('click', function(e) {
    e.preventDefault();
    window.location.hash = '#all-topics';
    $('.menu li a').removeClass('active');
    $(this).addClass('active');
    showAllTopics();
  });

  // Add logout handler
  $('#logout-link').on('click', function(e) {
    e.preventDefault();
    localStorage.removeItem('user');
    $.get('/api/logout', function(resp) {
      if (resp && resp.redirect) {
        window.location.href = resp.redirect;
      }
    });
  });

  // --- Login state helpers ---
  window.isLogin = function() {
    return !!localStorage.getItem('user');
  };
  window.getUserInfo = function() {
    var user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  };

  // --- Profile Modal Logic ---
  $('#profile-link').on('click', function(e) {
    e.preventDefault();
    var user = window.getUserInfo();
    if (!user) {
      window.showModalOverlay('<div style="padding:2em;">User info not found.</div>');
      return;
    }
    var html = '';
    html += '<div style="min-width:340px;max-width:440px;background:var(--primary-white);border-radius:14px;box-shadow:0 2px 12px var(--shadow-purple);padding:32px 28px 24px 28px;">';
    html += '<h2 style="margin-top:0;color:var(--primary-purple);font-size:1.5em;">Profile</h2>';
    html += '<div style="margin-bottom:10px;"><b style="color:var(--primary-purple);">Username:</b> <span style="color:#222;">' + user.username + '</span></div>';
    html += '<div style="margin-bottom:18px;"><b style="color:var(--primary-purple);">Name:</b> <span style="color:#222;">' + user.name + '</span></div>';
    if (user.group_details && user.group_details.length > 0) {
      html += '<h3 style="margin-bottom:0.5em;margin-top:1.5em;color:var(--primary-purple);font-size:1.1em;">Groups</h3>';
      html += '<table style="width:100%;border-collapse:collapse;background:var(--primary-white);margin-bottom:10px;">';
      html += '<thead><tr>';
      html += '<th style="background:var(--accent-purple);color:var(--primary-white);font-weight:600;padding:10px 8px;text-align:left;border-radius:8px 0 0 0;">Group Name</th>';
      html += '<th style="background:var(--accent-purple);color:var(--primary-white);font-weight:600;padding:10px 8px;text-align:left;border-radius:0 8px 0 0;">Role</th>';
      html += '</tr></thead>';
      html += '<tbody>';
      user.group_details.forEach(function(g, idx) {
        html += '<tr style="border-bottom:1px solid var(--border-purple);' + (idx === user.group_details.length-1 ? 'border-bottom:none;' : '') + '">';
        html += '<td style="padding:10px 8px;color:#222;">' + g.group_name + '</td>';
        html += '<td style="padding:10px 8px;color:#222;">' + g.role + '</td>';
        html += '</tr>';
      });
      html += '</tbody></table>';
    }
    html += '<div style="text-align:right;margin-top:1.5em;"><button id="close-profile-modal" style="padding:7px 22px;background:var(--accent-purple);color:var(--primary-white);border:none;border-radius:8px;font-size:1em;cursor:pointer;">Close</button></div>';
    html += '</div>';
    window.showModalOverlay(html);
    $(document).off('click', '#close-profile-modal').on('click', '#close-profile-modal', function() {
      window.hideModalOverlay();
    });
  });
});