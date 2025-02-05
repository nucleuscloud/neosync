---
title: RBAC
description: Learn how to use RBAC to manage user permissions in Neosync
id: rbac
hide_title: false
slug: /guides/rbac
---

## Introduction

RBAC (Role-Based Access Control) is a system that allows you to manage user permissions in Neosync.

## Neosync Version Availability

RBAC is available for all team and enterprise accounts in Neosync Cloud.

For OSS users, RBAC is only available with a valid Enterprise license.

## How to configure RBAC

RBAC is configured in the Neosync UI under the `Settings` page in the `Members` tab.

When inviting a new user to your team, the role of the user will be selected during the invitation process.

![Invite User](/img/rbac/invite-user.png)

After a user has accepted their invitation and they become a member, the role of the user can be changed in the `Members` tab.

> If you incorrectly set the role during the invite, the invite may be removed and a new invite may be sent with the correct role.

In the member's table, click the three dots on the right side of the row and select `Update Role`.

You'll be presented with a role update form modal where the role may be updated.

> Only admins may update the role of users.

![Update Role](/img/rbac/update-role.png)

## Roles

The following roles are available for configuration:

### Account Admin

This is the most permissive role for an account and grants full access to all features and settings within the account in Neosync.

This role is currently the only role that may create Account API Keys. This is because today Account API Keys are not granular and effectively operate as account admins.

### Job Developer

This role is used to manage jobs within an account. It has access to view basic account information.

This role has permission to manage jobs and connections within the account.

### Job Executor

This role is used to execute jobs within an account. It has access to view basic account information.

This role has permission view connections and jobs within the account. It may also trigger job runs.

### Job Viewer

This role is used to view jobs within an account. It has access to view basic account information.

This role has permission to view jobs and connections within the account, but is not able to trigger job runs.
