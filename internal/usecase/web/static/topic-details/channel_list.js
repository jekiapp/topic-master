let currentTopic = '';
let nsqdHosts = [];

function updateChannelsTable(topic, hosts) {
    if (!topic || !hosts || hosts.length === 0) return;

    const params = new URLSearchParams({
        topic: topic,
        hosts: hosts.join(',')
    });

    fetch(`/api/topic/nsq/list-channels?${params}`)
        .then(response => response.json())
        .then(resp => {
            const data = resp.data;
            const tbody = document.getElementById('channels-table-body');
            tbody.innerHTML = '';

            data.channels.forEach(channel => {
                const row = document.createElement('tr');
                
                // Name column
                const nameCell = document.createElement('td');
                nameCell.textContent = channel.name;
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
                    value.textContent = state.value;
                    
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
                    value.textContent = message.value;
                    
                    messageItem.appendChild(label);
                    messageItem.appendChild(value);
                    messagesCell.appendChild(messageItem);
                });
                row.appendChild(messagesCell);

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