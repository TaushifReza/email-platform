import { inviteUser, listUsers, updateUser } from "@/lib/api";

export const usersService = {
  list: listUsers,
  invite: inviteUser,
  update: updateUser,
};