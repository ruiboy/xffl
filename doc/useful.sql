-- ffl player season w/ round in/out
select ap.name, ac.name, fc.name, frfrom.name from_round, frto.name to_round
from ffl.player_season fps
         left outer join ffl.club_season fcs on fcs.id = fps.club_season_id
         left outer join ffl.club fc on fc.id = fcs.club_id
         left outer join afl.player_season aps on aps.id = fps.afl_player_season_id
         left outer join afl.player ap on ap.id = aps.player_id
         left outer join afl.club_season acs on acs.id = aps.club_season_id
         left outer join afl.club ac on ac.id = acs.club_id
         left outer join ffl.round frfrom on frfrom.id = fps.from_round_id
         left outer join ffl.round frto on frto.id = fps.to_round_id
order by fc.name, ap.name;

-- ffl teams
select fr.name as round,
       fc.name as ffl_club,
       ap.name as afl_player,
       ac.name as afl_club,
       fpm.position,
       am.data_status as afl_match_st,
       (fpm.afl_player_match_id is not null) as pm_afl_linked,
       fpm.status as pm_ffl_st,
       fpm.drv_afl_status as pm_afl_st,
       fpm.drv_score
from ffl.round fr
         join ffl.match fm on fm.round_id = fr.id
         join ffl.club_match fcm on fcm.match_id = fm.id
         join ffl.club_season fcs on fcs.id = fcm.club_season_id
         join ffl.club fc on fc.id = fcs.club_id
         join ffl.player_match fpm on fpm.club_match_id = fcm.id
         join ffl.player_season fps on fps.id = fpm.player_season_id
         left join afl.player_season aps on aps.id = fps.afl_player_season_id
         left join afl.player ap on ap.id = aps.player_id
         left join afl.club_season acs on acs.id = aps.club_season_id
         left join afl.club ac on ac.id = acs.club_id
         left join afl.club_match acm
                   on acm.club_season_id = acs.id
                       and acm.match_id in (select id from afl.match where round_id =
                                                                           fr.afl_round_id)
         left join afl.match am on am.id = acm.match_id
         left join afl.player_match apm
                   on apm.club_match_id = acm.id
                       and apm.player_season_id = aps.id
where fr.name = 'Round 4'
  and fc.name = 'Ruiboys'
order by fc.name, fpm.position, ap.name;

-- reset player match status
BEGIN;
UPDATE ffl.player_match
SET status = 'named', drv_score = null
  WHERE id IN (
      SELECT fpm.id
      FROM ffl.player_match fpm
      JOIN ffl.club_match fcm ON fcm.id = fpm.club_match_id
      JOIN ffl.match fm ON fm.id = fcm.match_id
      JOIN ffl.round fr ON fr.id = fm.round_id
      JOIN ffl.club_season fcs ON fcs.id = fcm.club_season_id
      JOIN ffl.club fc ON fc.id = fcs.club_id
      WHERE fr.name = 'Round 4'
        AND fc.name = 'Ruiboys'
  );
ROLLBACK;