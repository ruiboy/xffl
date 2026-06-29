-- ffl player tenure for a season
select ap.name, ac.name, fc.name, frfrom.name from_round, frto.name to_round, aps.id, acs.season_id
from ffl.player_season fps
         left outer join ffl.club_season fcs on fcs.id = fps.club_season_id
         left outer join ffl.club fc on fc.id = fcs.club_id
         left outer join afl.player_season aps on aps.id = fps.afl_player_season_id
         left outer join afl.player ap on ap.id = aps.player_id
         left outer join afl.club_season acs on acs.id = aps.club_season_id
         left outer join afl.club ac on ac.id = acs.club_id
R         left outer join ffl.round frfrom on frfrom.id = fps.from_round_id
         left outer join ffl.round frto on frto.id = fps.to_round_id
order by fc.name, ap.name;


-- ffl team for a round; compares stored and inferred status
with afl_data as (
    -- Players with an AFL match in this round
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
                then null
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

    union all

    -- Players whose AFL club has a bye in this round (no club_match exists)
    select
        aps.id as aps_id,
        ap.name as afl_player_name,
        ac.name as afl_club_name,
        ab.round_id as afl_round_id,
        null as am_data_status,
        null as apm_id,
        false as apm_exists,
        'bye' as apm_status_inferred
    from afl.player_season aps
             join afl.player ap on ap.id = aps.player_id
             join afl.club_season acs on acs.id = aps.club_season_id
             join afl.club ac on ac.id = acs.club_id
             join afl.bye ab on ab.club_season_id = acs.id
)
select
    fr.name as ffl_round,
    fc.name as ffl_club,
    fcm.id as ffl_club_match_id,
    fpm.id as ffl_player_match_id,
    ad.afl_player_name as afl_player,
    ad.afl_club_name as afl_club,
    fpm.position as ffl_position,
    fpm.display_order,
    fpm.backup_positions,
    fpm.interchange_position,
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
where fr.name = 'Round 12'
  and fc.name = 'The Howling Cows'
order by fc.name, fpm.position, fpm.display_order, ad.afl_player_name;


-- afl match scores per club for a round, one row per club;
-- compares computed_score against the stored drv_score
with target_round as (
    select r.id, r.name as round_name
    from afl.round r
    join afl.season s on s.id = r.season_id
    join afl.league l on l.id = s.league_id
    where l.name = 'AFL'
      and s.name = 'AFL 2026'
      and r.name = 'Opening Round'
),
player_totals as (
    select
        pm.club_match_id,
        sum(pm.goals)   as goals,
        sum(pm.behinds) as behinds
    from afl.player_match pm
    group by pm.club_match_id
),
club_scores as (
    select
        cm.match_id,
        cm.side,
        c.name                    as club_name,
        coalesce(pt.goals, 0)     as player_goals,
        coalesce(pt.behinds, 0)   as player_behinds,
        cm.rushed_behinds         as rushed_behinds,
        coalesce(pt.goals, 0) * 6
            + coalesce(pt.behinds, 0)
            + cm.rushed_behinds   as computed_score,
        cm.drv_score              as drv_score
    from afl.club_match cm
    join afl.club_season cs on cs.id = cm.club_season_id
    join afl.club c         on c.id = cs.club_id
    left join player_totals pt on pt.club_match_id = cm.id
)
select
    tr.round_name,
    m.id        as match_id,
    cs.side,
    cs.club_name,
    cs.player_goals,
    cs.player_behinds,
    cs.rushed_behinds,
    cs.computed_score,
    cs.drv_score,
    m.start_dt
from afl.match m
join target_round tr on tr.id = m.round_id
join club_scores cs  on cs.match_id = m.id
order by m.start_dt, cs.side desc;


-- afl match stats for a player
SELECT pm.id, ps.id, m.round_id, p.name, s.name, r.name,
       pm.kicks, pm.handballs, pm.marks, pm.hitouts, pm.tackles, pm.goals,
       (pm.kicks + pm.handballs + 2*pm.marks + 4*pm.tackles + 5*pm.goals) as star,
       pm.deleted_at, m.data_status
FROM afl.player_match pm
         JOIN afl.club_match cm ON cm.id = pm.club_match_id
         JOIN afl.match m ON m.id = cm.match_id
         JOIN afl.round r ON m.round_id = r.id
         JOIN afl.season s on r.season_id = s.id
         JOIN afl.player_season ps ON ps.id = pm.player_season_id
         JOIN afl.player p ON p.id = ps.player_id
WHERE p.name = 'Dan Houston'
ORDER BY ps.id, m.round_id;


-- FFL match results for a season (one row per match, one col per club)
SELECT
    r.name                                                        AS round,
    CASE WHEN 'Cheetahs' IN (home_c.name, away_c.name)
         THEN CASE WHEN home_c.name = 'Cheetahs' THEN home_cm.drv_score ELSE away_cm.drv_score END
    END                                                           AS cheetahs,
    CASE WHEN 'Cheetahs' IN (home_c.name, away_c.name)
         THEN CASE WHEN home_c.name = 'Cheetahs' THEN home_cm.drv_premiership_points ELSE away_cm.drv_premiership_points END
    END                                                           AS cheetahs_pts,
    CASE WHEN 'Ruiboys' IN (home_c.name, away_c.name)
         THEN CASE WHEN home_c.name = 'Ruiboys' THEN home_cm.drv_score ELSE away_cm.drv_score END
    END                                                           AS ruiboys,
    CASE WHEN 'Ruiboys' IN (home_c.name, away_c.name)
         THEN CASE WHEN home_c.name = 'Ruiboys' THEN home_cm.drv_premiership_points ELSE away_cm.drv_premiership_points END
    END                                                           AS ruiboys_pts,
    CASE WHEN 'Slashers' IN (home_c.name, away_c.name)
         THEN CASE WHEN home_c.name = 'Slashers' THEN home_cm.drv_score ELSE away_cm.drv_score END
    END                                                           AS slashers,
    CASE WHEN 'Slashers' IN (home_c.name, away_c.name)
         THEN CASE WHEN home_c.name = 'Slashers' THEN home_cm.drv_premiership_points ELSE away_cm.drv_premiership_points END
    END                                                           AS slashers_pts,
    CASE WHEN 'The Howling Cows' IN (home_c.name, away_c.name)
         THEN CASE WHEN home_c.name = 'The Howling Cows' THEN home_cm.drv_score ELSE away_cm.drv_score END
    END                                                           AS thc,
    CASE WHEN 'The Howling Cows' IN (home_c.name, away_c.name)
         THEN CASE WHEN home_c.name = 'The Howling Cows' THEN home_cm.drv_premiership_points ELSE away_cm.drv_premiership_points END
    END                                                           AS thc_pts
FROM ffl.match m
    JOIN ffl.round r              ON r.id = m.round_id
    JOIN ffl.season s             ON s.id = r.season_id
    JOIN afl.season afl_s         ON afl_s.id = s.afl_season_id
    JOIN ffl.club_match home_cm   ON home_cm.match_id = m.id AND home_cm.side = 'home'
    JOIN ffl.club_season home_cs  ON home_cs.id = home_cm.club_season_id
    JOIN ffl.club home_c          ON home_c.id = home_cs.club_id
    JOIN ffl.club_match away_cm   ON away_cm.match_id = m.id AND away_cm.side = 'away'
    JOIN ffl.club_season away_cs  ON away_cs.id = away_cm.club_season_id
    JOIN ffl.club away_c          ON away_c.id = away_cs.club_id
WHERE afl_s.name = 'AFL 2026'
  AND m.deleted_at IS NULL
  AND r.deleted_at IS NULL
ORDER BY r.afl_round_id, m.id;

