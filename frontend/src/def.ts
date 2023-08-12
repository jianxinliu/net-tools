export enum PingType {
    PING = 1,
    MTR = 2
}

export class MtrRow {
    IP = ""
    Sent = 0
    Recv = 0
    Min = 0
    Max = 0
    Loss = 0.0
    Last = 0
}