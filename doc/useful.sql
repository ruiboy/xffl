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

-- ffl team
with afl_data as (
    select
        aps.id as aps_id,
        ap.name as afl_player_name,
        ac.name as afl_club_name,
        am.round_id as afl_round_id,
        am.data_status as am_data_status,
        apm.id as apm_id,
        (apm.id is not null) as apm_exists,
        case
            when (am.data_status is null or am.data_status = 'no_data') and apm.id is null
                then null
            when (am.data_status is null or am.data_status = 'no_data') and apm.id is not null
                then 'named'
            when am.data_status = 'partial' and apm.id is null
                then 'dnp'
            when am.data_status = 'partial' and apm.id is not null
                then 'playing'
            when am.data_status = 'final' and apm.id is null
                then 'dnp'
            when am.data_status = 'final' and apm.id is not null
                then 'played'
            end as apm_status_inferred
    from afl.player_season aps
             left join afl.player ap on ap.id = aps.player_id
             left join afl.club_season acs on acs.id = aps.club_season_id
             left join afl.club ac on ac.id = acs.club_id
             left join afl.club_match acm on acm.club_season_id = acs.id
             left join afl.match am on am.id = acm.match_id
             left join afl.player_match apm
                       on apm.club_match_id = acm.id
                           and apm.player_season_id = aps.id
)
select
    fr.name as ffl_round,
    fc.name as ffl_club,
    ad.afl_player_name as afl_player,
    ad.afl_club_name as afl_club,
    fpm.position as ffl_position,
    ad.am_data_status,
    ad.apm_exists,
    ad.apm_status_inferred,
    (fpm.afl_player_match_id is not null) as fpm_apm_linked,
    fpm.drv_afl_status as fpm_drv_afl_status,
    fpm.status as fpm_status,
    fpm.drv_score
from ffl.round fr
         join ffl.match fm on fm.round_id = fr.id
         join ffl.club_match fcm on fcm.match_id = fm.id
         join ffl.club_season fcs on fcs.id = fcm.club_season_id
         join ffl.club fc on fc.id = fcs.club_id
         join ffl.player_match fpm on fpm.club_match_id = fcm.id
         join ffl.player_season fps on fps.id = fpm.player_season_id
         left join afl_data ad on ad.aps_id = fps.afl_player_season_id and ad.afl_round_id = fr.afl_round_id
where fr.name = 'Round 10'
  and fc.name = 'Ruiboys'
order by fc.name, fpm.position, ad.afl_player_name;

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