let currentTopic = '';
let nsqdHosts = [];
let currentTopicDetail = null;

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
                const nameWrapper = document.createElement('div');
                nameWrapper.className = 'channel-name-wrapper';

                // Channel name row (name + paused label)
                const nameRow = document.createElement('div');
                nameRow.className = 'channel-name-row';

                const nameSpan = document.createElement('span');
                nameSpan.textContent = channel.name;
                nameSpan.className = 'channel-name-bold';
                nameRow.appendChild(nameSpan);

                if (channel.is_paused) {
                    const pausedLabel = document.createElement('span');
                    pausedLabel.textContent = 'paused';
                    pausedLabel.className = 'channel-paused-label';
                    nameRow.appendChild(pausedLabel);
                }

                // Add claim link next to channel name
                const claimLink = document.createElement('a');
                claimLink.href = 'javascript:void(0)';
                claimLink.className = 'claim-link';
                claimLink.style.marginLeft = '10px';
                claimLink.textContent = 'Claim';
                claimLink.onclick = function(e) {
                    e.preventDefault();
                    e.stopPropagation();
                    window.showClaimModal(channel.id, channel.name, window.handleClaimEntity);
                };
                nameRow.appendChild(claimLink);

                nameWrapper.appendChild(nameRow);

                if (channel.group_owner) {
                    const ownerDiv = document.createElement('div');
                    ownerDiv.className = 'channel-group-owner channel-sub-title';
                    ownerDiv.textContent = 'Owner: ' + channel.group_owner;
                    nameWrapper.appendChild(ownerDiv);
                }

                // Add consumer count section
                const consumerDiv = document.createElement('div');
                consumerDiv.className = 'channel-consumer-count channel-sub-title ';
                consumerDiv.textContent = 'Consumers: -';
                nameWrapper.appendChild(consumerDiv);

                nameCell.appendChild(nameWrapper);
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

                // Other action buttons (pause/resume, delete, empty)
                // Only show pause or resume based on is_paused
                if (channel.is_paused) {
                    const resumeBtn = document.createElement('button');
                    resumeBtn.className = 'action-icon-btn btn-resume';
                    resumeBtn.title = channel.is_free_action ? 'Resume Channel' : 'You do not have permission.';
                    resumeBtn.textContent = '▶️';
                    resumeBtn.disabled = false;
                    resumeBtn.onclick = (e) => {
                        e.stopPropagation();
                        checkChannelActionPermissionAsync(
                            channel.is_free_action,
                            channel.group_owner,
                            'pause',
                            channel.id,
                            function(allowed) {
                                if (!allowed) return;
                                handleChannelAction('resume', channel.name);
                            }
                        );
                    };
                    actionsWrapper.appendChild(resumeBtn);
                } else {
                    const pauseBtn = document.createElement('button');
                    pauseBtn.className = 'action-icon-btn btn-pause';
                    pauseBtn.title = channel.is_free_action ? 'Pause Channel' : 'You do not have permission.';
                    pauseBtn.textContent = '⏸️';
                    pauseBtn.disabled = false;
                    pauseBtn.onclick = (e) => {
                        e.stopPropagation();
                        checkChannelActionPermissionAsync(
                            channel.is_free_action,
                            channel.group_owner,
                            'pause',
                            channel.id,
                            function(allowed) {
                                if (!allowed) return;
                                handleChannelAction('pause', channel.name);
                            }
                        );
                    };
                    actionsWrapper.appendChild(pauseBtn);
                }
                // Delete and empty buttons
                const deleteBtn = document.createElement('button');
                deleteBtn.className = 'action-icon-btn btn-delete';
                deleteBtn.title = channel.is_free_action ? 'Delete Channel' : 'You do not have permission.';
                deleteBtn.textContent = '🗑️';
                deleteBtn.disabled = false;
                deleteBtn.onclick = (e) => {
                    e.stopPropagation();
                    checkChannelActionPermissionAsync(
                        channel.is_free_action,
                        channel.group_owner,
                        'delete',
                        channel.id,
                        function(allowed) {
                            if (!allowed) return;
                            handleChannelAction('delete', channel.name);
                        }
                    );
                };
                actionsWrapper.appendChild(deleteBtn);
                const emptyBtn = document.createElement('button');
                emptyBtn.className = 'action-icon-btn btn-empty';
                emptyBtn.title = channel.is_free_action ? 'Empty Channel' : 'You do not have permission.';
                emptyBtn.textContent = '🧹';
                emptyBtn.disabled = false;
                emptyBtn.onclick = (e) => {
                    e.stopPropagation();
                    checkChannelActionPermissionAsync(
                        channel.is_free_action,
                        channel.group_owner,
                        'empty',
                        channel.id,
                        function(allowed) {
                            if (!allowed) return;
                            handleChannelAction('empty', channel.name);
                        }
                    );
                };
                actionsWrapper.appendChild(emptyBtn);
                
                actionsCell.appendChild(actionsWrapper);
                row.appendChild(actionsCell);

                tbody.appendChild(row);
            });
            // Fetch topic stats using hosts and topic name
            if (window.fetchAndUpdateStats && window.currentTopicDetail) {
                window.fetchAndUpdateStats(window.currentTopicDetail);
            }
        })
        .catch(error => console.error('Error fetching channels:', error));
}

function refreshChannels(detail) {
    currentTopic = detail.name;
    nsqdHosts = detail.nsqd_hosts;
    currentTopicDetail = detail;
    updateChannelsTable(detail.name, detail.nsqd_hosts);
}

function handleChannelAction(action, channelName) {
    console.log(`${action} action clicked for channel: ${channelName}`);
    // Find the row and get the channel id
    const row = document.querySelector(`tr[data-channel-name="${channelName}"]`);
    const channelId = row ? row.getAttribute('data-channel-id') : null;
    function showModal(msg, onConfirm, confirmText = 'OK', cancelText = null) {
        let modalHtml = `<div style="text-align:center;">
            <div style="font-size:1.1em;margin-bottom:18px;">${msg}</div>`;
        if (onConfirm) {
            modalHtml += `<button id="modal-confirm-btn" style="margin-right:18px;padding:8px 18px;background:#ff2d2d;color:#fff;border:none;border-radius:6px;font-weight:600;cursor:pointer;">${confirmText}</button>`;
            if (cancelText) {
                modalHtml += `<button id="modal-cancel-btn" style="padding:8px 18px;background:#eee;color:#333;border:none;border-radius:6px;font-weight:600;cursor:pointer;">${cancelText}</button>`;
            }
        } else {
            modalHtml += `<button id="modal-ok-btn" style="padding:8px 18px;background:#eee;color:#333;border:none;border-radius:6px;font-weight:600;cursor:pointer;">OK</button>`;
        }
        modalHtml += '</div>';
        if (window.parent && window.parent.showModalOverlay) {
            window.parent.showModalOverlay(modalHtml);
            setTimeout(function() {
                if (onConfirm) {
                    const doc = window.parent.document;
                    doc.getElementById('modal-confirm-btn')?.addEventListener('click', function() {
                        window.parent.hideModalOverlay();
                        onConfirm();
                    });
                    if (cancelText) {
                        doc.getElementById('modal-cancel-btn')?.addEventListener('click', function() {
                            window.parent.hideModalOverlay();
                        });
                    }
                } else {
                    window.parent.document.getElementById('modal-ok-btn')?.addEventListener('click', function() {
                        window.parent.hideModalOverlay();
                    });
                }
            }, 0);
        } else {
            if (onConfirm) {
                if (confirm(`Fallback: ${msg}`)) onConfirm();
            } else {
                alert(msg);
            }
        }
    }
    switch(action) {
        case 'bookmark':
            showModal('Bookmark channel: ' + channelName);
            break;
        case 'pause':
            if (!channelId) {
                showModal('Channel ID not found.');
                return;
            }
            showModal(`Are you sure you want to pause channel: ${channelName}?`, function() {
                const btn = row.querySelector('.btn-pause');
                if (btn) btn.disabled = true;
                fetch(`/api/channel/nsq/pause?id=${encodeURIComponent(channelId)}&channel=${encodeURIComponent(channelName)}&entity_id=${encodeURIComponent(channelId)}`)
                    .then(resp => resp.json())
                    .then(data => {
                        refreshChannels(currentTopicDetail);
                        if (window.fetchAndUpdateStats && window.currentTopicDetail) {
                            window.fetchAndUpdateStats(window.currentTopicDetail);
                        }
                    })
                    .catch(err => {
                        showModal('Failed to pause channel');
                    })
                    .finally(() => {
                        if (btn) btn.disabled = false;
                    });
            }, 'Yes, Pause', 'Cancel');
            break;
        case 'resume':
            if (!channelId) {
                showModal('Channel ID not found.');
                return;
            }
            showModal(`Are you sure you want to resume channel: ${channelName}?`, function() {
                const btn = row.querySelector('.btn-resume');
                if (btn) btn.disabled = true;
                fetch(`/api/channel/nsq/resume?id=${encodeURIComponent(channelId)}&channel=${encodeURIComponent(channelName)}&entity_id=${encodeURIComponent(channelId)}`)
                    .then(resp => resp.json())
                    .then(data => {
                        refreshChannels(currentTopicDetail);
                        if (window.fetchAndUpdateStats && window.currentTopicDetail) {
                            window.fetchAndUpdateStats(window.currentTopicDetail);
                        }
                    })
                    .catch(err => {
                        showModal('Failed to resume channel');
                    })
                    .finally(() => {
                        if (btn) btn.disabled = false;
                    });
            }, 'Yes, Resume', 'Cancel');
            break;
        case 'delete':
            if (!channelId) {
                showModal('Channel ID not found.');
                return;
            }
            showModal(`Are you sure you want to delete channel: ${channelName}? This cannot be undone.`, function() {
                const btn = row.querySelector('.btn-delete');
                if (btn) btn.disabled = true;
                fetch(`/api/channel/nsq/delete?id=${encodeURIComponent(channelId)}&entity_id=${encodeURIComponent(channelId)}`)
                    .then(resp => resp.json())
                    .then(data => {
                        refreshChannels(currentTopicDetail);
                        if (window.fetchAndUpdateStats && window.currentTopicDetail) {
                            window.fetchAndUpdateStats(window.currentTopicDetail);
                        }
                    })
                    .catch(err => {
                        showModal('Failed to delete channel');
                    })
                    .finally(() => {
                        if (btn) btn.disabled = false;
                    });
            }, 'Yes, Delete', 'Cancel');
            break;
        case 'empty':
            if (!channelId) {
                showModal('Channel ID not found.');
                return;
            }
            showModal(`Are you sure you want to empty channel: ${channelName}? This cannot be undone.`, function() {
                const btn = row.querySelector('.btn-empty');
                if (btn) btn.disabled = true;
                fetch(`/api/channel/nsq/empty?id=${encodeURIComponent(channelId)}&channel=${encodeURIComponent(channelName)}&entity_id=${encodeURIComponent(channelId)}`)
                    .then(resp => resp.json())
                    .then(data => {
                        refreshChannels(currentTopicDetail);
                        if (window.fetchAndUpdateStats && window.currentTopicDetail) {
                            window.fetchAndUpdateStats(window.currentTopicDetail);
                        }
                    })
                    .catch(err => {
                        showModal('Failed to empty channel');
                    })
                    .finally(() => {
                        if (btn) btn.disabled = false;
                    });
            }, 'Yes, Empty', 'Cancel');
            break;
        default:
            console.log(`Unknown action: ${action}`);
    }
}

// Helper to check channel action permission (async, like topic_details.js)
function checkChannelActionPermissionAsync(canFlag, groupOwner, actionName, entityId, cb) {
    if (canFlag) { cb(true); return; }
    if (!(window.parent.isLogin && window.parent.isLogin())) {
        showChannelPermissionModal({id: entityId}, actionName, groupOwner);
        cb(false); return;
    }
    // User is logged in, check with backend
    fetch('/api/auth/check-action', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action: actionName, entity_id: entityId })
    })
    .then(resp => resp.json())
    .then(data => {
        if (data && data.data && data.data.allowed) {
            cb(true);
        } else {
            showChannelPermissionModal({id: entityId}, actionName, groupOwner);
            cb(false);
        }
    })
    .catch(() => {
        showChannelPermissionModal({id: entityId}, actionName, groupOwner);
        cb(false);
    });
}

// Update showChannelPermissionModal to accept groupOwner for login message
function showChannelPermissionModal(channel, action, groupOwner) {
    const applyUrl = `#tickets-new?type=channel_action&entity_id=${channel.id}&action=${action}`;
    let msg = '';
    if (!(window.parent.isLogin && window.parent.isLogin())) {
        msg = `This channel is owned by ${groupOwner || '-'}.<br/>You must login to perform this action.`;
    } else {
        msg = `You do not have permission to perform this action.<br/><br/><a href="${applyUrl}" target="_blank">Apply for permission</a>`;
    }
    if (window.parent && window.parent.showModalOverlay) {
        window.parent.showModalOverlay(msg);
    } else {
        alert('You do not have permission. Apply at: ' + applyUrl);
    }
} 