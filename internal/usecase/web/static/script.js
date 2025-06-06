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
        $('#user-dropdown').show();
        $('#login-signup-link').hide();
        $('#user-dropdown .user-name').text(userData.name + ' ▼');
      } else {
        $('#user-dropdown').hide();
        $('#login-signup-link').show();
      }
      // Hide Access Control if not root
      if (!userData || userData.root === false) {
        $(".menu li a").filter(function() {
          return $(this).text().trim() === 'Access Control';
        }).closest('li').hide();
      }
    }).fail(function(jqxhr) {
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
    mainIframe.attr('src', 'my_topics/index.html');
  }

  function showAccessControl() {
    mainIframe.attr('src', 'acl/index.html');
  }

  const allTopicsMenu = $('.menu li a').filter(function() {
    return $(this).text().trim() === 'All Topics';
  });

  function showAllTopics() {
    mainIframe.attr('src', 'all_topics/index.html');
  }

  function setActiveMenuByHash(hash) {
    $('.menu li a').removeClass('active');
    if (hash === '#access') {
      accessControlMenu.addClass('active');
    } else if (hash === '#tickets' || hash.startsWith('#ticket-detail')) {
      ticketsMenu.addClass('active');
    } else if (hash === '#all_topics') {
      allTopicsMenu.addClass('active');
    } else {
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
    } else if (hash.startsWith('#ticket-detail')) {
      showTicketDetail();
    } else if (hash === '#all_topics') {
      showAllTopics();
    } else {
      showMyTopics();
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
    window.location.hash = '#all_topics';
    $('.menu li a').removeClass('active');
    $(this).addClass('active');
    showAllTopics();
  });

  // Add logout handler
  $('#logout-link').on('click', function(e) {
    e.preventDefault();
    $.get('/api/logout', function(resp) {
      if (resp && resp.redirect) {
        window.location.href = resp.redirect;
      }
    });
  });
});