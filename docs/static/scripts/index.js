document.addEventListener('DOMContentLoaded', function () {
  const isIndex =
    window.location.pathname === '/' ||
    window.location.pathname.endsWith('/index.html');

  if (isIndex) {
    // Remove elements with class 'sidebar-container'
    document.querySelectorAll('.sidebar-container').forEach(function (el) {
      el.remove();
    });
    // Remove elements with class 'hextra-toc'
    document.querySelectorAll('.hextra-toc').forEach(function (el) {
      el.remove();
    });
  }
});
