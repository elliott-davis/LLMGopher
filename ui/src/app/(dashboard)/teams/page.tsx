import { fetchTeams } from '@/components/teams/teams-data';
import { TeamsClient } from '@/components/teams/teams-client';

export default async function TeamsPage() {
  const { teams, unavailable } = await fetchTeams();
  return <TeamsClient teams={teams} unavailable={unavailable} />;
}
