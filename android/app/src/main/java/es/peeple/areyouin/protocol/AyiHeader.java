package es.peeple.areyouin.protocol;

public class AyiHeader {
    // Modifiers
    public static final byte M_CREATE_EVENT = 0x00;
    public static final byte M_CANCEL_EVENT = 0x01;
    public static final byte M_INVITE_USERS = 0x02;
    public static final byte M_CANCEL_USERS_INVITATION = 0x03;
    public static final byte M_CONFIRM_ATTENDANCE = 0x04;
    public static final byte M_MODIFY_EVENT_DATE = 0x05;
    public static final byte M_MODIFY_EVENT_MESSAGE = 0x06;
    public static final byte M_MODIFY_EVENT = 0x07;
    public static final byte M_VOTE_CHANGE = 0x08;
    public static final byte M_USER_POSITION = 0x09;
    public static final byte M_USER_POSITION_RANGE = 0x0A;
    public static final byte M_USER_CREATE_ACCOUNT = 0x0B;
    public static final byte M_USER_NEW_AUTH_TOKEN = 0x0C;
    public static final byte M_USER_AUTH= 0x0D;

    // Notifications
    public static final byte M_EVENT_CREATED = 0x40;
    public static final byte M_EVENT_CANCELLED = 0x41;
    public static final byte M_EVENT_EXPIRED = 0x42;
    public static final byte M_EVENT_DATE_MODIFIED = 0x43;
    public static final byte M_EVENT_MESSAGE_MODIFIED = 0x44;
    public static final byte M_EVENT_MODIFIED = 0x45;
    public static final byte M_INVITATION_RECEIVED = 0x46;
    public static final byte M_INVITATION_CANCELLED = 0x47;
    public static final byte M_ATTENDANCE_STATUS = 0x48;
    public static final byte M_EVENT_CHANGE_DATE_PROPOSED = 0x49;
    public static final byte M_EVENT_CHANGE_MESSAGE_PROPOSED = 0x4A;
    public static final byte M_EVENT_CHANGE_PROPOSED = 0x4B;
    public static final byte M_VOTING_STATUS = 0x4C;
    public static final byte M_VOTING_FINISHED = 0x4D;
    public static final byte M_CHANGE_ACCEPTED = 0x4E;
    public static final byte M_CHANGE_DISCARDED = 0x4F;
    public static final byte M_ACCESS_GRANTED = 0x50;
    public static final byte M_OK = 0x7E;
    public static final byte M_ERROR = 0x7F;

    // Requests
    public static final byte M_PING = (byte) 0x80;
    public static final byte M_READ_EVENT = (byte) 0x81;
    public static final byte M_LIST_AUTHORED_EVENTS = (byte) 0x82;
    public static final byte M_LIST_PRIVATE_EVENTS = (byte) 0x83;
    public static final byte M_LIST_PUBLIC_EVENTS = (byte) 0x84;
    public static final byte M_HISTORY_AUTHORED_EVENTS = (byte) 0x85;
    public static final byte M_HISTORY_PRIVATE_EVENTS = (byte) 0x86;
    public static final byte M_HISTORY_PUBLIC_EVENTS = (byte) 0x87;

    // Responses
    public static final byte M_PONG = (byte) 0xC0;
    public static final byte M_EVENT_INFO = (byte) 0xC1;
    public static final byte M_EVENTS_LIST = (byte) 0xC2;


    public byte version;
    public short token;
    public byte type;
    public short size;

    public AyiHeader() {
        version = 0;
        token = 0;
        type = AyiHeader.M_ERROR;
        size = 6;
    }
}
