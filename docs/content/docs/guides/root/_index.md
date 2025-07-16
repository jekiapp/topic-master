+++
weight = 320
date = '2025-07-10T09:12:28+07:00'
draft = true
title = 'Root'
+++

Root group and root user by default is created when Topic Master initialized with a new data.
This group is a main administrator for Topic Master.
By default, root group has one `root` user, but it can also have more than one user. 
New user for root group can be created in the Access Control page, or anyone can signup and apply to be a member of root group (if approved).
root group has no admin role.

#User Group
the first thing to do for root user is creating groups. It is important because new user would looking for their group when they try to signup.
Root user could also creating new users, the password would be auto-generated, then it can be given to the target person then they will be forced to create a new password for the first login.

#Tickets assignment
Root group member(s) will be assigned as approver for signup application, along with the admin of the group being applied to.
It also assigned for an application targeting to a group that has no admin role in it.
For example, if someone try to apply to publish to a topic that owned by GroupA, but there's no user with admin role registered in GroupA, then the root group member will be assigned as approver.

