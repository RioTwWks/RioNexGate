export interface Invite {
  id: number;
  user_id: number;
  token: string;
  label?: string;
  max_uses: number;
  used_count: number;
  expires_at?: string | null;
  revoked_at?: string | null;
  usable: boolean;
  created_at: string;
}

export interface CreateInviteInput {
  label?: string;
  max_uses?: number;
  expires_hours?: number;
}
