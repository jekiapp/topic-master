function getQueryParam(name) {
    const url = new URL(window.location.href);
    return url.searchParams.get(name);
}

function renderObject(obj, $container) {
    if (!obj) return;
    const $table = $('<table></table>');
    $.each(obj, function(key, value) {
        const $row = $('<tr></tr>');
        $row.append($('<th></th>').text(key));
        $row.append($('<td></td>').text(value));
        $table.append($row);
    });
    $container.empty().append($table);
}

function renderArray(arr, $container) {
    if (!arr || arr.length === 0) {
        $container.text('No data.');
        return;
    }
    const $table = $('<table></table>');
    const $thead = $('<thead></thead>');
    const $headerRow = $('<tr></tr>');
    $.each(Object.keys(arr[0]), function(_, key) {
        $headerRow.append($('<th></th>').text(key));
    });
    $thead.append($headerRow);
    $table.append($thead);
    const $tbody = $('<tbody></tbody>');
    $.each(arr, function(_, item) {
        const $row = $('<tr></tr>');
        $.each(item, function(_, val) {
            $row.append($('<td></td>').text(val));
        });
        $tbody.append($row);
    });
    $table.append($tbody);
    $container.empty().append($table);
}

$(function() {
    const id = getQueryParam('id');
    if (!id) {
        $('#application-section').text('No application ID provided.');
        return;
    }
    $.get(`/api/signup/application`, { id: id })
        .done(function(data) {
            renderObject(data.application, $('#application-section'));
            renderObject(data.user, $('#user-section'));
            renderArray(data.assignments, $('#assignments-section'));
            renderArray(data.histories, $('#histories-section'));
        })
        .fail(function() {
            $('#application-section').text('Failed to load application detail.');
        });
}); 