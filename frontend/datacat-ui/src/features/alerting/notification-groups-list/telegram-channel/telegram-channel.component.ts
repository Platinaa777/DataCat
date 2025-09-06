import { Component,Input } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { PanelModule } from 'primeng/panel';
import { NotificationChannel,TelegramSettings } from '../../../../entities';

@Component({
  standalone: true,
  selector: 'datacat-telegram-channel',
  templateUrl: './telegram-channel.component.html',
  styleUrl: './telegram-channel.component.scss',
  imports: [PanelModule, ButtonModule, FormsModule],
})
export class TelegramChannelComponent {
  @Input() public channel?: NotificationChannel;

  protected settingsAsTelegramSettings(): TelegramSettings {
    return this.channel?.settings as TelegramSettings;
  }
}
