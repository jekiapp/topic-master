// Modal overlay logic using jQuery
(function() {
    function ensureModal() {
        if ($('#modal-overlay').length) return;
        const $modal = $('<div id="modal-overlay"><div class="modal-content" id="modal-content"></div></div>');
        $modal.css('display', 'none');
        $('body').append($modal);
        $modal.on('click', function(e) {
            if (e.target === $modal[0]) hideModalOverlay();
        });
    }

    function showModalOverlay(content) {
        ensureModal();
        const $modal = $('#modal-overlay');
        const $modalContent = $('#modal-content');
        $modalContent.html(content);
        $modal.css('display', 'flex');
    }

    function hideModalOverlay() {
        $('#modal-overlay').css('display', 'none');
    }

    // Expose to global for iframe access
    window.showModalOverlay = showModalOverlay;
    window.hideModalOverlay = hideModalOverlay;
})(); 