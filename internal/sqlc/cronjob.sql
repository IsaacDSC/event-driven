CREATE EXTENSION pg_cron;


CREATE OR REPLACE FUNCTION clean_old_events() RETURNS void AS $$
DECLARE
last_month DATE := (CURRENT_DATE - INTERVAL '1 month')::date;
    first_day_of_last_month DATE := date_trunc('month', last_month);
    first_day_of_this_month DATE := date_trunc('month', CURRENT_DATE);
BEGIN

DELETE FROM tx_sagas
WHERE created_at >= first_day_of_last_month AND created_at < first_day_of_this_month;

DELETE FROM transactions
WHERE created_at >= first_day_of_last_month AND created_at < first_day_of_this_month;

RAISE NOTICE 'Successfully cleaned events from % to %', first_day_of_last_month, first_day_of_this_month;

END;
$$ LANGUAGE plpgsql;

SELECT cron.schedule('0 0 * * *', 'SELECT clean_old_events();');