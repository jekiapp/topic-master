+++
weight = 330
date = '2025-07-10T09:12:22+07:00'
draft = true
title = 'Login'
+++

A logged-in user has access to all the basic features available to non-logged-in users, plus additional features that require a user session.

This page introduces the term "Entity," which refers to an object that access control is applied to, such as a topic or channel.

## Bookmark

Logged-in users can bookmark entities. All bookmarked topics will appear in the "My Topics" menu. Currently, bookmarking a channel only displays a mark icon.

## Claim

Entities can be claimed by users. There are two types of entities:
- **Unowned entities**: Entities that do not currently have an owner (Group owner: None). When claiming an unowned entity, the default approver is the group admin of the group the entity is being claimed into.
- **Group-owned entities**: Entities that are already owned by a specific group. External users (those not in the owning group) can request to claim these entities, but the approver will be the group admin of the current owner group.

Once an entity is claimed by a group, it is protected from users outside the owning group. By default, all actions are restricted for external users.

## Apply for Permission

If an action on an entity is restricted because the user is not a member of the owning group, the user can apply for access by clicking the relevant action button (e.g., publish, tail). If the action is prohibited, a popup will appear with a link to open a new application page. On this page, the user can select the permissions they wish to request. The approver for these applications is the admin of the owning group.

## Ticket Assignment

If a user is a root member or has an admin role in a group, the ticket list page will show assignments for that user. All applications that the user is eligible to approve are listed there. Only one assignee is required to take action; if someone else has already acted, the other assignees will be marked as passed.
