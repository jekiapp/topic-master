$(function() {
  const iframeContainer = $('.iframe-container');
  const myTopicsMenu = $('.menu li a').filter(function() {
    return $(this).text().trim() === 'My Topics';
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

  myTopicsMenu.on('click', function(e) {
    e.preventDefault();
    $('.menu li a').removeClass('active');
    $(this).addClass('active');
    showMyTopics();
  });

  // Optionally, load My Topics by default on page load
  showMyTopics();
}); 