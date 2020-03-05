package main

frame_complete := false

func Start_frame()  {
	frame_complete = false;
}

func is_frame_complete() bool
{
    return frame_complete;
}

func set_CRT(bool interlaced, int mode, bool frame_mode)
{
    reg.set_CRT(interlaced, mode, frame_mode);

    GSMessagePayload payload;
    payload.crt_payload = { interlaced, mode, frame_mode };

    gs_thread.send_message({ GSCommand::set_crt_t, payload });
}