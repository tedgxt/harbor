import {
  Component,
  OnInit,
  OnChanges,
  Input,
  ViewChild,
  Output,
  EventEmitter,
  SimpleChanges
} from "@angular/core";
import { Webhook, Target } from "../webhook";
import { NgForm } from "@angular/forms";
import {ClrLoadingState} from "@clr/angular";
import { finalize } from "rxjs/operators";
import { WebhookService } from "../webhook.service";
import { MessageHandlerService } from "../../../shared/message-handler/message-handler.service";

@Component({
  selector: 'add-webhook-form',
  templateUrl: './add-webhook-form.component.html',
  styleUrls: ['./add-webhook-form.component.scss']
})
export class AddWebhookFormComponent implements OnInit, OnChanges {
  closable: boolean = true;
  staticBackdrop: boolean = true;
  isCheckValid: boolean = false;
  isEndpointValid: boolean = false;
  checking: boolean = false;
  checkBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
  webhookForm: NgForm;
  submitting: boolean = false;
  webhookTarget: Target = new Target();

  @Input() projectId: number;
  @Input() webhook: Webhook;
  @Input() isModify: boolean;
  @Input() isOpen: boolean;
  @Output() edit = new EventEmitter<boolean>();
  @Output() close = new EventEmitter<boolean>();
  @ViewChild("webhookForm") currentForm: NgForm;


  constructor(
    private webhookService: WebhookService,
    private messageHandlerService: MessageHandlerService
  ) { }

  ngOnInit() {
    
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes['isOpen'] && changes['isOpen'].currentValue) { 
      Object.assign(this.webhookTarget, this.webhook.targets[0]);
    }
  }

  onTestEndpoint() {
    this.checkBtnState = ClrLoadingState.LOADING;
    this.checking = true;

    this.webhookService
      .testEndpoint(this.projectId, {
        targets: [this.webhookTarget]
      })
      .pipe(finalize(() => (this.checking = false)))
      .subscribe(
        response => {
          this.checkBtnState = ClrLoadingState.SUCCESS;
          this.isCheckValid = true;
        },
        error => {
          this.checkBtnState = ClrLoadingState.DEFAULT;
          this.isCheckValid = false;
          this.messageHandlerService.handleError(error);
        }
      );
  }

  onCancel() {
    this.close.emit(false);
    this.currentForm.reset();
  }

  onSubmit() {
    const rx = this.isModify 
      ? this.webhookService.editWebhook(this.projectId, this.webhook.id, Object.assign(this.webhook, { targets: [this.webhookTarget] }))
      : this.webhookService.createWebhook(this.projectId, {
        targets: [this.webhookTarget],
        hook_types: ["downloadChart", "deleteChart", "uploadChart", "deleteImage", "pullImage", "pushImage", "scanningFailed", "scanningCompleted"],
        enabled: true,
      }); 
    rx.pipe(finalize(() => (this.submitting = false)))
      .subscribe(
        response => {
          this.edit.emit(this.isModify);
        },
        error => {
          this.messageHandlerService.handleError(error);
        }
      );
  }

  setCertValue($event: any): void {
    this.webhookTarget.skip_cert_verify = !$event;
  }

  public get isValid(): boolean {
    return (
      this.currentForm &&
      this.currentForm.valid &&
      !this.submitting &&
      !this.checking
    );
  }
}
