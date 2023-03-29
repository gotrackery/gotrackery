CREATE OR REPLACE FUNCTION insert_position(
    p_protocol VARCHAR(128), p_deviceid VARCHAR(128), p_servertime timestamp, p_devicetime timestamp,
    p_fixtime timestamp, p_valid BOOLEAN, p_latitude DOUBLE PRECISION, p_longitude DOUBLE PRECISION,
    p_altitude DOUBLE PRECISION, p_speed DOUBLE PRECISION, p_course DOUBLE PRECISION, p_address VARCHAR(512),
    p_attributes VARCHAR(4000), p_accuracy DOUBLE PRECISION, p_network VARCHAR(4000))
    RETURNS TABLE (device_id INTEGER)
    LANGUAGE plpgsql
AS
$$
DECLARE
    tc_device_id INTEGER;
    tc_position_id BIGINT;
BEGIN
    SELECT id INTO tc_device_id FROM tc_devices WHERE uniqueid = p_deviceid;
    IF NOT FOUND THEN
        INSERT INTO tc_devices (name, uniqueid, groupid)
        VALUES (p_deviceid,p_deviceid,1)
        RETURNING id INTO tc_device_id;
    END IF;

    INSERT INTO tc_positions (
        protocol, deviceid, servertime, devicetime,
        fixtime, valid, latitude, longitude,
        altitude, speed, course, address,
        attributes, accuracy, network)
    VALUES (p_protocol, tc_device_id, p_servertime, p_devicetime,
            p_fixtime, p_valid, p_latitude, p_longitude,
            p_altitude, p_speed, p_course, p_address,
            p_attributes, p_accuracy, p_network)
    RETURNING id INTO tc_position_id;

    UPDATE tc_devices
    SET positionid = tc_position_id,
        lastupdate = now(),
        status = 'online'
    WHERE id = tc_device_id;

    RETURN QUERY
        SELECT tc_device_id;
END;
$$;