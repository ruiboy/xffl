-- ffl player ins/outs
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
where (fps.from_round_id is not null and frfrom.name != 'Round 1')
or fps.to_round_id is not null;