$(function() {
  const iframeContainer = $('.iframe-container');
  const myTopicsMenu = $('.menu li a').filter(function() {
    return $(this).text().trim() === 'My Topics';
  });
  const accessControlMenu = $('.menu li a').filter(function() {
    return $(this).text().trim() === 'Access Control';
  });

  function showMyTopics() {
    let iframe = iframeContainer.find('iframe');
    if (iframe.length === 0) {
      iframe = $('<iframe>', {
        src: 'my_topics/index.html',
        style: 'width:100%;height:100%;border:none;min-height:600px;',
        frameborder: 0,
        allowfullscreen: true
      });
      iframeContainer.empty().append(iframe);
    } else if (iframe.attr('src') !== 'my_topics/index.html') {
      iframe.attr('src', 'my_topics/index.html');
    }
  }

  function showAccessControl() {
    let iframe = iframeContainer.find('iframe');
    if (iframe.length === 0) {
      iframe = $('<iframe>', {
        src: 'acl/index.html',
        style: 'width:100%;height:100%;border:none;min-height:600px;',
        frameborder: 0,
        allowfullscreen: true
      });
      iframeContainer.empty().append(iframe);
    } else if (iframe.attr('src') !== 'acl/index.html') {
      iframe.attr('src', 'acl/index.html');
    }
  }

  function setActiveMenuByHash(hash) {
    $('.menu li a').removeClass('active');
    if (hash === '#access') {
      accessControlMenu.addClass('active');
    } else {
      myTopicsMenu.addClass('active');
    }
  }

  function handleHashChange() {
    const hash = window.location.hash;
    setActiveMenuByHash(hash);
    if (hash === '#access') {
      showAccessControl();
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