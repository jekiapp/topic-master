+++
weight = 330
date = '2025-07-10T09:12:22+07:00'
draft = true
title = 'Login'
+++

For a logged-in user, it will have the all the basic features granted for non-login users, with additional feature that requiring user session.
In this page will be introduced a term "Entity" this is an object to apply the access control into, for example topic and channel

## Bookmark
Logged in user could bookmark the entity. For all the bookmarked topic, it will shown in the "My Topics" menu. For now, bookmarking a channel will only give a mark icon.

## Claim
Exntities is available to claim, whether it's a None ownered entities or a group ownered one. 
For orphaned entity, the default approver would be the group admin of the entity claimed into.
For claimed entity (entity with owned by specific group), external user still able to request to claim the entity, 
but the approver would be the group admin of the current owner.
Claimed entities will be protected for other user outside the group owner. All the action is restricted by default for external user.

## Apply for permission
If the action of an entity is restricted due to the user is not a member of the group owner, user can apply the access by clicking the action button (e.g publish, tail button).
Then if the action is prohibited, there should be a popup with a link to open a new application page. 
In the application page, user could check permissions that deemed suitable. 
The approver of the application page would be the admin of the group owner. 

## Ticket assignment
If a user is a root member or an admin role of a group, The ticket list page will show the assignment for the user. 
All the application that the user eligible to approve is listed there. 
Only one of the assignee required to take an action, so if someone else already took the action, the other assignee will be marked as pass.
