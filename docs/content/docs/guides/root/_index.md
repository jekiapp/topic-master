+++
weight = 320
date = '2025-07-10T09:12:28+07:00'
draft = true
title = 'Root User'
+++

The root group and root user are created by default when Topic Master is initialized with new data. This group serves as the main administrator for Topic Master. By default, the root group has one `root` user, but additional users can be added. New users for the root group can be created on the Access Control page, or anyone can sign up and apply to join the root group (subject to approval). The root group does not have an admin role.

## User Group

The first task for a root user is to create groups. This is important because new users will look for their group when signing up. Root users can also create new users; the password will be auto-generated and can be given to the intended person, who will then be required to create a new password upon their first login.
<br/>

<div style="display: flex; justify-content: center; align-items: center; margin: 2rem 0;">
  <img src="/images/docs/user group.png" alt="User group" style="max-width: 700px; width: 100%; border-radius: 1rem; box-shadow: 0 4px 16px rgba(0,0,0,0.08);" />
</div>

## Ticket Assignment

Root group members are assigned as approvers for signup applications, along with the admin of the group being applied to. They are also assigned to applications targeting a group that does not have an admin role.

For example, if someone applies to publish to a topic owned by `Group A`, but there is no user with an admin role registered in `Group A`, then the root group members will be assigned as approvers.
