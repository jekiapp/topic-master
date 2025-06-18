let currentTopic = '';
let nsqdHosts = [];

function updateChannelsTable(topic, hosts) {
    if (!topic || !hosts || hosts.length === 0) return;

    // Extract address if hosts is array of objects
    const hostAddresses = hosts.map(host => {
        if (typeof host === 'object' && host.address) {
            return host.address;
        }
        return host;
    });

    const params = new URLSearchParams({
        topic: topic,
        hosts: hostAddresses.join(',')
    });

    fetch(`/api/topic/nsq/list-channels?${params}`)
        .then(response => response.json())
        .then(resp => {
            const data = resp.data;
            const tbody = document.getElementById('channels-table-body');
            tbody.innerHTML = '';

            data.channels.forEach(channel => {
                const row = document.createElement('tr');
                row.setAttribute('data-channel-name', channel.name);
                row.setAttribute('data-channel-id', channel.id);
                
                // Name column
                const nameCell = document.createElement('td');
                nameCell.className = 'channel-name-cell';

                // Channel name text
                const nameSpan = document.createElement('span');
                nameSpan.textContent = channel.name;
                nameCell.appendChild(nameSpan);
                row.appendChild(nameCell);

                // States column
                const statesCell = document.createElement('td');
                statesCell.className = 'states-cell';
                
                const states = [
                    { label: 'In Flight', value: channel.in_flight },
                    { label: 'Requeued', value: channel.requeued },
                    { label: 'Deferred', value: channel.deferred }
                ];

                states.forEach(state => {
                    const stateItem = document.createElement('div');
                    stateItem.className = 'state-item';
                    
                    const label = document.createElement('span');
                    label.className = 'state-label';
                    label.textContent = state.label;
                    
                    const value = document.createElement('span');
                    value.className = 'state-value';
                    value.textContent = Number(state.value).toLocaleString();
                    
                    stateItem.appendChild(label);
                    stateItem.appendChild(value);
                    statesCell.appendChild(stateItem);
                });
                row.appendChild(statesCell);

                // Messages column
                const messagesCell = document.createElement('td');
                messagesCell.className = 'messages-cell';
                
                const messages = [
                    { label: 'Depth', value: channel.depth },
                    { label: 'Total', value: channel.messages }
                ];

                messages.forEach(message => {
                    const messageItem = document.createElement('div');
                    messageItem.className = 'message-item';
                    
                    const label = document.createElement('span');
                    label.className = 'message-label';
                    label.textContent = message.label;
                    
                    const value = document.createElement('span');
                    value.className = 'message-value';
                    value.textContent = Number(message.value).toLocaleString();
                    
                    messageItem.appendChild(label);
                    messageItem.appendChild(value);
                    messagesCell.appendChild(messageItem);
                });
                row.appendChild(messagesCell);

                // Actions column
                const actionsCell = document.createElement('td');
                actionsCell.className = 'actions-cell';
                
                // Create wrapper div for buttons
                const actionsWrapper = document.createElement('div');
                actionsWrapper.className = 'actions-wrapper';
                
                // --- Bookmark PNG icon as action button ---
                const bookmarkImg = document.createElement('img');
                bookmarkImg.className = 'bookmark-img-channel';
                bookmarkImg.style.width = '20px';
                bookmarkImg.style.height = '20px';
                bookmarkImg.style.verticalAlign = 'middle';
                bookmarkImg.style.marginRight = '8px';
                bookmarkImg.src = channel.is_bookmarked ? '/icons/bookmark-true.png' : '/icons/bookmark-false.png';
                bookmarkImg.title = channel.is_bookmarked ? 'Remove Bookmark' : 'Add Bookmark';
                bookmarkImg.style.cursor = (window.parent.isLogin && window.parent.isLogin()) ? 'pointer' : 'not-allowed';
                bookmarkImg.onclick = function(e) {
                    e.stopPropagation();
                    if (!(window.parent.isLogin && window.parent.isLogin())) {
                        alert('Please log in to bookmark channels.');
                        return;
                    }
                    const newState = !channel.is_bookmarked;
                    // Optimistic UI update
                    bookmarkImg.src = newState ? '/icons/bookmark-true.png' : '/icons/bookmark-false.png';
                    bookmarkImg.title = newState ? 'Remove Bookmark' : 'Add Bookmark';
                    channel.is_bookmarked = newState;
                    fetch('/api/entity/toggle-bookmark', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ entity_id: channel.id, bookmark: newState })
                    }).then(resp => {
                        if (!resp.ok) {
                            // Revert UI if failed
                            bookmarkImg.src = !newState ? '/icons/bookmark-true.png' : '/icons/bookmark-false.png';
                            bookmarkImg.title = !newState ? 'Remove Bookmark' : 'Add Bookmark';
                            channel.is_bookmarked = !newState;
                            resp.json().then(data => {
                                alert(data.error || 'Failed to toggle bookmark');
                            });
                        }
                    }).catch(() => {
                        // Revert UI if failed
                        bookmarkImg.src = !newState ? '/icons/bookmark-true.png' : '/icons/bookmark-false.png';
                        bookmarkImg.title = !newState ? 'Remove Bookmark' : 'Add Bookmark';
                        channel.is_bookmarked = !newState;
                        alert('Failed to toggle bookmark');
                    });
                };
                actionsWrapper.appendChild(bookmarkImg);

                // Other action buttons (pause, delete, empty)
                const actionButtons = [
                    { name: 'pause', title: 'Pause Channel', icon: '⏸️' },
                    { name: 'delete', title: 'Delete Channel', icon: '🗑️' },
                    { name: 'empty', title: 'Empty Channel', icon: '🧹' }
                ];

                actionButtons.forEach(action => {
                    const button = document.createElement('button');
                    button.className = `action-icon-btn btn-${action.name}`;
                    button.title = action.title;
                    button.textContent = action.icon;
                    button.onclick = () => handleChannelAction(action.name, channel.name);
                    actionsWrapper.appendChild(button);
                });
                
                actionsCell.appendChild(actionsWrapper);
                row.appendChild(actionsCell);

                tbody.appendChild(row);
            });
        })
        .catch(error => console.error('Error fetching channels:', error));
}

function refreshChannels(topic, hosts) {
    currentTopic = topic;
    nsqdHosts = hosts;
    updateChannelsTable(topic, hosts);
}

function handleChannelAction(action, channelName) {
    console.log(`${action} action clicked for channel: ${channelName}`);
    // TODO: Implement actual functionality for each action
    switch(action) {
        case 'bookmark':
            alert(`Bookmark channel: ${channelName}`);
            break;
        case 'pause':
            alert(`Pause channel: ${channelName}`);
            break;
        case 'delete':
            if (confirm(`Are you sure you want to delete channel: ${channelName}?`)) {
                alert(`Delete channel: ${channelName}`);
            }
            break;
        case 'empty':
            if (confirm(`Are you sure you want to empty channel: ${channelName}?`)) {
                alert(`Empty channel: ${channelName}`);
            }
            break;
        default:
            console.log(`Unknown action: ${action}`);
    }
} 