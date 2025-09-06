import { Component,Input } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { PanelModule } from 'primeng/panel';
import { EmailSettings,NotificationChannel } from '../../../../entities';

@Component({
  standalone: true,
  selector: 'datacat-email-channel',
  templateUrl: './email-channel.component.html',
  styleUrl: './email-channel.component.scss',
  imports: [PanelModule, ButtonModule, FormsModule],
})
export class EmailChannelComponent {
  @Input() public channel?: NotificationChannel;

  protected settingsAsEmailSettings(): EmailSettings {
    return this.channel?.settings as EmailSettings;
  }
}
